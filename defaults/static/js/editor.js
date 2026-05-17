(function () {
	var currentFile = "";
	var originalContent = "";
	var isEditing = false;
	var isDirty = false;
	var pollTimer = null;
	var reloadChangedHandler = null;
	var debounceTimer = null;
	var DEBOUNCE_DELAY = 150;
	var LARGE_DOC_THRESHOLD = 5000;

	function encodePath(path) {
		return path.split("/").map(encodeURIComponent).join("/");
	}

	function init() {
		var body = document.body;
		currentFile = (body.getAttribute("data-current-file") || "").trim();
		if (!currentFile || !currentFile.toLowerCase().endsWith(".md")) {
			return;
		}

		var editBtn = document.querySelector(".editor-btn-edit");
		if (editBtn) {
			editBtn.addEventListener("click", enterEditMode);
		}

		var saveBtn = document.querySelector(".editor-btn-save");
		if (saveBtn) {
			saveBtn.addEventListener("click", saveContent);
		}

		var cancelBtn = document.querySelector(".editor-btn-cancel");
		if (cancelBtn) {
			cancelBtn.addEventListener("click", cancelEdit);
		}

		var previewBtn = document.querySelector(".editor-btn-preview");
		if (previewBtn) {
			previewBtn.addEventListener("click", togglePreview);
		}

		document.addEventListener("keydown", handleKeydown);

		window.addEventListener("beforeunload", function (e) {
			if (isDirty) {
				e.preventDefault();
			}
		});

		interceptSidebarLinks();

		reloadChangedHandler = handleExternalReload;
		window.addEventListener("reload-changed", reloadChangedHandler);

		restoreScrollPosition();
	}

	function enterEditMode() {
		fetch("/api/raw/" + encodePath(currentFile))
			.then(function (resp) {
				if (!resp.ok) {
					throw new Error("Failed to load file");
				}
				return resp.text();
			})
			.then(function (text) {
				originalContent = text;
				var textarea = document.querySelector(".editor-textarea");
				textarea.value = text;
				isEditing = true;
				isDirty = false;
				document.body.setAttribute("data-editing", "true");
				toggleUI(true);
				textarea.focus();

				textarea.oninput = function () {
					isDirty = textarea.value !== originalContent;
					scheduleRender();
				};

				textarea.setSelectionRange(0, 0);
				document.documentElement.scrollTop = 0;

				var previewBtn = document.querySelector(".editor-btn-preview");
				if (previewBtn) previewBtn.classList.add("active");

				DEBOUNCE_DELAY = 150;

				var wrapper = document.querySelector(".editor-split-wrapper");
				if (wrapper) wrapper.classList.remove("no-preview");
				renderPreview();

				pollTimer = setInterval(checkExternalChanges, 5000);
			})
			.catch(function (err) {
				alert("Failed to load file: " + err.message);
			});
	}

	function renderPreview() {
		var preview = document.querySelector(".editor-split-preview");
		var textarea = document.querySelector(".editor-textarea");
		if (!preview || !textarea) return;
		if (typeof marked === "undefined") return;

		// XSS boundary: marked.parse() passes through raw HTML by default.
		// Since textarea content is self-authored (local dev tool),
		// this is an acceptable self-XSS boundary.
		// Relative image paths resolve against the page URL, not the
		// file directory, so they may appear broken in preview.
		preview.innerHTML = marked.parse(textarea.value);
	}

	function scheduleRender() {
		clearTimeout(debounceTimer);
		var textarea = document.querySelector(".editor-textarea");
		var lines = textarea ? textarea.value.split("\n").length : 0;
		var delay = lines > LARGE_DOC_THRESHOLD ? 300 : DEBOUNCE_DELAY;
		debounceTimer = setTimeout(function () {
			requestAnimationFrame(renderPreview);
		}, delay);
	}

	function saveContent() {
		var textarea = document.querySelector(".editor-textarea");
		var content = textarea.value;
		var saveBtn = document.querySelector(".editor-btn-save");
		saveBtn.disabled = true;
		saveBtn.textContent = "Saving...";

		fetch("/api/edit/" + encodePath(currentFile), {
			method: "POST",
			body: content,
		})
			.then(function (resp) {
				if (!resp.ok) {
					return resp.json().then(function (data) {
						throw new Error(data.error || "Save failed");
					});
				}
				return resp.json();
			})
			.then(function () {
				originalContent = content;
				isDirty = false;
				saveSidebarState();
				showToast("Saved", "success");
				fetch("/api/raw/" + encodePath(currentFile))
					.then(function(resp) { return resp.text(); })
					.then(function(raw) {
						if (typeof marked !== "undefined") {
							var rendered = marked.parse(raw);
							var previewContent = document.querySelector(".preview-content");
							if (previewContent) {
								previewContent.innerHTML = rendered;
							}
							var splitPreview = document.querySelector(".editor-split-preview");
							if (splitPreview) {
								splitPreview.innerHTML = rendered;
							}
						}
					})
					.catch(function() {});
			})
			.catch(function (err) {
				saveBtn.disabled = false;
				saveBtn.textContent = "Save";
				showToast(err.message, "error");
			});
	}

	function togglePreview() {
		if (!isEditing) return;

		var wrapper = document.querySelector(".editor-split-wrapper");
		var previewBtn = document.querySelector(".editor-btn-preview");
		if (!wrapper) return;

		var isHidden = wrapper.classList.toggle("no-preview");

		if (!isHidden) {
			renderPreview();
		}

		if (previewBtn) {
			previewBtn.classList.toggle("active", !isHidden);
		}
	}

	function cancelEdit() {
		if (isDirty) {
			if (!confirm("You have unsaved changes. Discard them?")) {
				return;
			}
		}
		exitEditMode();
	}

	function exitEditMode() {
		isEditing = false;
		isDirty = false;
		document.body.removeAttribute("data-editing");
		toggleUI(false);
		if (pollTimer) {
			clearInterval(pollTimer);
			pollTimer = null;
		}
		if (reloadChangedHandler) {
			window.removeEventListener("reload-changed", reloadChangedHandler);
			reloadChangedHandler = null;
		}
		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
		var preview = document.querySelector(".editor-split-preview");
		if (preview) preview.innerHTML = "";
		var wrapper = document.querySelector(".editor-split-wrapper");
		if (wrapper) wrapper.classList.remove("no-preview");
		var previewBtn = document.querySelector(".editor-btn-preview");
		if (previewBtn) previewBtn.classList.remove("active");
	}

	function checkExternalChanges() {
		fetch("/api/raw/" + encodePath(currentFile))
			.then(function (resp) {
				if (!resp.ok) {
					if (confirm("This file has been deleted externally. Close the editor?")) {
						exitEditMode();
						window.location.reload();
					}
					return Promise.reject(null);
				}
				return resp.text();
			})
			.then(function (text) {
				if (text === originalContent) return;
				if (confirm("This file has been modified externally. Overwrite with your changes or reload the latest version?")) {
					originalContent = text;
				} else {
					var textarea = document.querySelector(".editor-textarea");
					textarea.value = text;
					originalContent = text;
					isDirty = false;
				}
			})
			.catch(function () {});
	}

	function toggleUI(editMode) {
		var previewToolbar = document.querySelector(".preview-toolbar");
		var editorToolbar = document.querySelector(".editor-toolbar");
		var editorContainer = document.querySelector(".editor-container");
		var previewContent = document.querySelector(".preview-content");

		if (previewToolbar) previewToolbar.style.display = editMode ? "none" : "";
		if (editorToolbar) editorToolbar.style.display = editMode ? "" : "none";
		if (editorContainer) editorContainer.style.display = editMode ? "" : "none";
		if (previewContent) previewContent.style.display = editMode ? "none" : "";
	}

	function handleExternalReload(e) {
		var changedFile = e.detail.file;
		if (changedFile === currentFile) {
			showToast("Saved", "success");
		} else {
			showToast("File updated: " + changedFile, "success");
		}
	}

	function handleKeydown(e) {
		if (!isEditing) return;

		var isCtrl = e.ctrlKey || e.metaKey;

		if (isCtrl && e.key === "s") {
			e.preventDefault();
			saveContent();
			return;
		}

		if (isCtrl && e.key === "Enter") {
			e.preventDefault();
			saveContent();
			return;
		}

		if (e.key === "Escape") {
			e.preventDefault();
			cancelEdit();
			return;
		}
	}

	function showToast(message, type) {
		var existing = document.querySelector(".editor-toast");
		if (existing) existing.remove();

		var toast = document.createElement("div");
		toast.className = "editor-toast editor-toast-" + type;
		toast.textContent = message;
		toast.setAttribute("role", "status");
		document.body.appendChild(toast);

		setTimeout(function () {
			if (toast.parentNode) toast.remove();
		}, type === "success" ? 1500 : 3000);
	}

	function interceptSidebarLinks() {
		document.addEventListener("click", function (e) {
			if (!isEditing) return;
			var target = e.target;
			while (target && target !== document.body) {
				if (target.tagName === "A" && target.closest(".docs-sidebar")) {
					e.preventDefault();
					if (isDirty) {
						showNavDialog(target.href);
					} else {
						exitEditMode();
						window.location.href = target.href;
					}
					return;
				}
				target = target.parentElement;
			}
		});
	}

	function showNavDialog(targetUrl) {
		var overlay = document.createElement("div");
		overlay.className = "editor-dialog-overlay";

		var dialog = document.createElement("div");
		dialog.className = "editor-dialog";

		var message = document.createElement("p");
		message.textContent = "You have unsaved changes. What would you like to do?";
		dialog.appendChild(message);

		var btnSave = document.createElement("button");
		btnSave.textContent = "Save and switch";
		btnSave.className = "editor-btn editor-btn-save";
		btnSave.addEventListener("click", function () {
			var textarea = document.querySelector(".editor-textarea");
			btnSave.disabled = true;
			btnSave.textContent = "Saving...";
			fetch("/api/edit/" + encodePath(currentFile), {
				method: "POST",
				body: textarea.value,
			}).then(function (resp) {
				if (!resp.ok) {
					return resp.json().then(function (data) {
						throw new Error(data.error || "Save failed");
					});
				}
				return resp.json();
			}).then(function () {
				isDirty = false;
				window.location.href = targetUrl;
			}).catch(function (err) {
				btnSave.disabled = false;
				btnSave.textContent = "Save and switch";
				overlay.querySelector("p").textContent =
					"Save failed: " + (err.message || "network error") + ". Try again?";
			});
		});

		var btnDiscard = document.createElement("button");
		btnDiscard.textContent = "Discard and switch";
		btnDiscard.className = "editor-btn";
		btnDiscard.addEventListener("click", function () {
			exitEditMode();
			window.location.href = targetUrl;
		});

		var btnCancel = document.createElement("button");
		btnCancel.textContent = "Continue editing";
		btnCancel.className = "editor-btn";
		btnCancel.addEventListener("click", function () {
			overlay.remove();
		});

		dialog.appendChild(btnSave);
		dialog.appendChild(btnDiscard);
		dialog.appendChild(btnCancel);
		overlay.appendChild(dialog);
		document.body.appendChild(overlay);
	}

	function saveSidebarState() {
		var details = document.querySelectorAll(".docs-sidebar details");
		var state = {};
		details.forEach(function (d, i) {
			var summary = d.querySelector("summary span.docs-sidebar-label");
			var key = summary ? summary.textContent : "section-" + i;
			state[key] = d.hasAttribute("open");
		});
		sessionStorage.setItem("go-grip-sidebar-state", JSON.stringify(state));
	}

	function restoreSidebarState() {
		var saved = sessionStorage.getItem("go-grip-sidebar-state");
		if (!saved) return;
		try {
			var state = JSON.parse(saved);
			var details = document.querySelectorAll(".docs-sidebar details");
			details.forEach(function (d, i) {
				var summary = d.querySelector("summary span.docs-sidebar-label");
				var key = summary ? summary.textContent : "section-" + i;
				if (state[key] === true) {
					d.setAttribute("open", "");
				}
			});
		} catch (_) {}
	}

	function restoreScrollPosition() {
		var saved = sessionStorage.getItem("go-grip-scrollTop");
		if (saved) {
			var top = parseInt(saved, 10);
			if (!isNaN(top) && top > 0) {
				setTimeout(function () {
					document.documentElement.scrollTop = top;
				}, 100);
			}
			sessionStorage.removeItem("go-grip-scrollTop");
		}
	}

	init();

	document.addEventListener("DOMContentLoaded", function () {
		restoreSidebarState();
	});
})();
