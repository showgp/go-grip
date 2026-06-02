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
	var isSyncingScroll = false;
	var pendingImages = [];

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

		var previewBtn = document.querySelector(".editor-btn-eye");
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

		// Import Image 按钮
		var importBtn = document.querySelector(".editor-btn-import");
		if (importBtn) {
			importBtn.addEventListener("click", function() {
				openImportDialog();
			});
		}

		// 对话框关闭按钮
		document.querySelector(".import-dialog-close")?.addEventListener("click", closeImportDialog);
		document.querySelector(".import-btn-cancel")?.addEventListener("click", closeImportDialog);

		// Tab 切换
		document.querySelectorAll(".import-tab").forEach(function(tab) {
			tab.addEventListener("click", function() {
				switchImportTab(this.dataset.tab);
			});
		});

		// 确认按钮
		document.querySelector(".import-btn-confirm")?.addEventListener("click", handleImportConfirm);

		// Pending Tray 按钮
		document.querySelector(".import-pending-insert-all")?.addEventListener("click", insertAllPending);
		document.querySelector(".import-pending-close")?.addEventListener("click", clearPendingTray);

		// 初始化和导入相关的交互
		initDropzone();
		initTextareaDrop();
		initPasteHandler();
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
				updateDoneBtnLabel();
				textarea.focus();

				textarea.oninput = function () {
					isDirty = textarea.value !== originalContent;
					updateDoneBtnLabel();
					scheduleRender();
				};

				textarea.setSelectionRange(0, 0);
				textarea.scrollTop = 0;
				document.documentElement.scrollTop = 0;
				requestAnimationFrame(function () {
					textarea.scrollTop = 0;
					textarea.setSelectionRange(0, 0);
				});

				var previewBtn = document.querySelector(".editor-btn-eye");
				if (previewBtn) previewBtn.classList.add("active");

				DEBOUNCE_DELAY = 150;

				var previewEl = document.querySelector(".editor-split-preview");
				textarea.addEventListener("scroll", syncScrollFromTextarea);
				if (previewEl) previewEl.addEventListener("scroll", syncScrollFromPreview);

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
		if (typeof marked === "undefined") {
			console.warn("marked not loaded; preview cannot render");
			return;
		}

		// XSS boundary: marked.parse() passes through raw HTML by default.
		// Since textarea content is self-authored (local dev tool),
		// this is an acceptable self-XSS boundary.
		// Relative image paths resolve against the page URL, not the
		// file directory, so they may appear broken in preview.
		preview.innerHTML = marked.parse(textarea.value);
		requestAnimationFrame(syncScrollFromTextarea);
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

	function syncScrollFromTextarea() {
		if (isSyncingScroll) return;
		isSyncingScroll = true;
		try {
			var textarea = document.querySelector(".editor-textarea");
			var preview = document.querySelector(".editor-split-preview");
			if (textarea && preview) {
				var th = textarea.scrollHeight - textarea.clientHeight;
				var ph = preview.scrollHeight - preview.clientHeight;
				preview.scrollTop = th > 0 ? (textarea.scrollTop / th) * ph : 0;
			}
		} finally {
			isSyncingScroll = false;
		}
	}

	function syncScrollFromPreview() {
		if (isSyncingScroll) return;
		isSyncingScroll = true;
		try {
			var textarea = document.querySelector(".editor-textarea");
			var preview = document.querySelector(".editor-split-preview");
			if (textarea && preview) {
				var ph = preview.scrollHeight - preview.clientHeight;
				var th = textarea.scrollHeight - textarea.clientHeight;
				textarea.scrollTop = ph > 0 ? (preview.scrollTop / ph) * th : 0;
			}
		} finally {
			isSyncingScroll = false;
		}
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
				saveBtn.disabled = false;
				saveBtn.textContent = "Save";
				originalContent = content;
				isDirty = false;
				updateDoneBtnLabel();
				saveSidebarState();
				showToast("Saved", "success");
				if (typeof marked !== "undefined") {
					var rendered = marked.parse(content);
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
			.catch(function (err) {
				saveBtn.disabled = false;
				saveBtn.textContent = "Save";
				showToast(err.message, "error");
				console.error("Save failed:", err);
			});
	}

	function togglePreview() {
		if (!isEditing) return;

		var wrapper = document.querySelector(".editor-split-wrapper");
		var previewBtn = document.querySelector(".editor-btn-eye");
		if (!wrapper) return;

		var isHidden = wrapper.classList.toggle("no-preview");

		if (!isHidden) {
			renderPreview();
		}

		if (previewBtn) {
			previewBtn.classList.toggle("active", !isHidden);
		}
	}

	function updateDoneBtnLabel() {
		var doneBtn = document.querySelector(".editor-btn-cancel");
		if (!doneBtn) return;
		doneBtn.textContent = isDirty ? "Cancel" : "Done";
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
		var textarea = document.querySelector(".editor-textarea");
		if (textarea) textarea.removeEventListener("scroll", syncScrollFromTextarea);
		var preview = document.querySelector(".editor-split-preview");
		if (preview) {
			preview.removeEventListener("scroll", syncScrollFromPreview);
			preview.innerHTML = "";
		}
		var wrapper = document.querySelector(".editor-split-wrapper");
		if (wrapper) wrapper.classList.remove("no-preview");
		var previewBtn = document.querySelector(".editor-btn-eye");
		if (previewBtn) previewBtn.classList.remove("active");

		pendingImages = [];
		var tray = document.querySelector(".import-pending-tray");
		if (tray) tray.style.display = "none";
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
				if (confirm("This file has been modified externally. Discard your changes and reload the latest version?")) {
					var textarea = document.querySelector(".editor-textarea");
					textarea.value = text;
					originalContent = text;
					isDirty = false;
					updateDoneBtnLabel();
					renderPreview();
				} else {
					originalContent = text;
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
		if (changedFile !== currentFile) {
			showToast("File updated: " + changedFile, "success");
			return;
		}
		if (!isEditing) return;
		fetch("/api/raw/" + encodePath(currentFile))
			.then(function (resp) {
				if (!resp.ok) throw new Error("Failed to fetch");
				return resp.text();
			})
			.then(function (text) {
				if (text === originalContent) return;
				if (isDirty && !confirm("This file has been modified externally. Discard your changes and reload?")) {
					return;
				}
				var textarea = document.querySelector(".editor-textarea");
				textarea.value = text;
				originalContent = text;
				isDirty = false;
				updateDoneBtnLabel();
				renderPreview();
				showToast("Reloaded latest version", "success");
			})
			.catch(function () {});
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

// ===== Image Import Functions =====

function openImportDialog() {
  var overlay = document.querySelector(".import-dialog-overlay");
  if (!overlay) return;
  overlay.style.display = "";
  resetDropzone();
  switchImportTab("local");
}

function closeImportDialog() {
  var overlay = document.querySelector(".import-dialog-overlay");
  if (!overlay) return;
  overlay.style.display = "none";
  var input = document.querySelector(".import-file-input");
  if (input) input.value = "";
  document.querySelectorAll(".import-preview-item").forEach(function(item) {
    if (item._objectUrl) URL.revokeObjectURL(item._objectUrl);
  });
}

function resetDropzone() {
  var list = document.querySelector(".import-preview-list");
  if (list) { list.style.display = "none"; list.innerHTML = ""; }
  var msg = document.querySelector(".import-dropzone-message");
  if (msg) msg.style.display = "";
}

function switchImportTab(tabName) {
  document.querySelectorAll(".import-tab").forEach(function(t) {
    t.classList.toggle("active", t.dataset.tab === tabName);
  });
  document.querySelectorAll(".import-tab-content").forEach(function(tc) {
    tc.classList.toggle("active", tc.dataset.tab === tabName);
  });
  var btn = document.querySelector(".import-btn-confirm");
  if (btn) {
    btn.textContent = tabName === "local" ? "Import" : "Insert";
  }
}

function handleImportConfirm() {
  var activeTab = document.querySelector(".import-tab.active");
  if (!activeTab) return;

  if (activeTab.dataset.tab === "url") {
    var urlInput = document.querySelector(".import-url-input");
    var url = urlInput ? urlInput.value.trim() : "";
    if (!url) {
      showToast("Please enter an image URL", "error");
      return;
    }
    insertTextAtCursor("![](" + url + ")\n");
    closeImportDialog();
    return;
  }

  var previewItems = document.querySelectorAll(".import-preview-item");
  if (previewItems.length === 0) {
    showToast("No images selected", "error");
    return;
  }

  var confirmBtn = this;
  confirmBtn.disabled = true;
  confirmBtn.textContent = "Importing...";

  var filesToImport = [];
  previewItems.forEach(function(item) {
    filesToImport.push({
      file: item._file,
      name: item._fileName
    });
  });

  if (filesToImport.length === 1) {
    var f = filesToImport[0];
    importSingleFile(f.file, f.name, function(path) {
      confirmBtn.disabled = false;
      confirmBtn.textContent = "Import";
      if (path) {
        closeImportDialog();
        var textarea = document.querySelector(".editor-textarea");
        if (textarea) {
          var pos = textarea.selectionStart;
          textarea.focus();
          textarea.selectionStart = textarea.selectionEnd = pos;
          insertTextAtCursor("![](" + encodeURI(path) + ")\n");
        }
      } else {
        showToast("Import failed", "error");
      }
    });
    return;
  }

  importAllFiles(filesToImport, function(successCount) {
    confirmBtn.disabled = false;
    confirmBtn.textContent = "Import";
    if (successCount > 0) {
      closeImportDialog();
    }
  });
}

function importAllFiles(files, callback) {
  var completed = 0;
  var successCount = 0;
  var total = files.length;
  var pendingAdditions = [];

  files.forEach(function(f) {
    importSingleFile(f.file, f.name, function(path, fileObj) {
      completed++;
      if (path) {
        successCount++;
        pendingAdditions.push({
          name: path.split("/").pop(),
          path: path,
          file: fileObj
        });
      }
      if (completed === total) {
        var readerCount = pendingAdditions.length;
        if (readerCount === 0) {
          if (callback) callback(successCount);
          return;
        }
        pendingAdditions.forEach(function(item) {
          var reader = new FileReader();
          reader.onload = function(e) {
            pendingImages.push({
              name: item.name,
              path: item.path,
              dataUrl: e.target.result
            });
            readerCount--;
            if (readerCount === 0) {
              updatePendingTray();
              if (callback) callback(successCount);
            }
          };
          reader.readAsDataURL(item.file);
        });
      }
    });
  });
}

function importSingleFile(file, fileName, callback) {
  var formData = new FormData();
  formData.append("file", file, fileName);

  var url = "/api/import?file=" + encodeURIComponent(currentFile);

  fetch(url, {
    method: "POST",
    body: formData
  })
  .then(function(resp) {
    if (!resp.ok) {
      return resp.json().then(function(data) {
        throw new Error(data.error || "Import failed");
      });
    }
    return resp.json();
  })
  .then(function(data) {
    callback(data.path, file);
  })
  .catch(function(err) {
    showToast("Import failed: " + err.message, "error");
    callback(null, null);
  });
}

function updatePendingTray() {
  var tray = document.querySelector(".import-pending-tray");
  var list = document.querySelector(".import-pending-list");
  var count = document.querySelector(".import-pending-count");
  if (!tray || !list) return;

  if (pendingImages.length === 0) {
    tray.style.display = "none";
    return;
  }

  tray.style.display = "";
  if (count) count.textContent = "Pending images (" + pendingImages.length + ")";

  list.innerHTML = "";
  pendingImages.forEach(function(img, index) {
    var item = document.createElement("div");
    item.className = "import-pending-item";
    item.title = img.name;

    var imgEl = document.createElement("img");
    imgEl.src = img.dataUrl;
    imgEl.alt = img.name;
    item.appendChild(imgEl);

    var nameEl = document.createElement("div");
    nameEl.className = "import-pending-item-name";
    nameEl.textContent = img.name;
    item.appendChild(nameEl);

    item.addEventListener("click", function() {
      insertImageAtIndex(index);
    });

    list.appendChild(item);
  });
}

function insertImageAtIndex(index) {
  if (index < 0 || index >= pendingImages.length) return;
  var img = pendingImages[index];
  insertTextAtCursor("![](" + encodeURI(img.path) + ")\n");
  pendingImages.splice(index, 1);
  updatePendingTray();
}

function insertAllPending() {
  if (pendingImages.length === 0) return;
  var text = pendingImages.map(function(img) {
    return "![](" + encodeURI(img.path) + ")";
  }).join("\n") + "\n";
  insertTextAtCursor(text);
  pendingImages = [];
  updatePendingTray();
}

function clearPendingTray() {
  pendingImages = [];
  updatePendingTray();
}

function insertTextAtCursor(text) {
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;

  var start = textarea.selectionStart;
  var end = textarea.selectionEnd;

  var before = textarea.value.substring(0, start);
  var after = textarea.value.substring(end);

  textarea.value = before + text + after;

  var newPos = start + text.length;
  textarea.setSelectionRange(newPos, newPos);
  textarea.focus();

  var event = new Event("input", { bubbles: true });
  textarea.dispatchEvent(event);
}

/* --- Dropzone (initialized once) --- */

var dropzoneInitialized = false;

function initDropzone() {
  if (dropzoneInitialized) return;
  var dropzone = document.querySelector(".import-dropzone");
  var fileInput = document.querySelector(".import-file-input");
  if (!dropzone || !fileInput) return;
  dropzoneInitialized = true;

  dropzone.addEventListener("click", function(e) {
    if (e.target.closest(".import-preview-item")) return;
    fileInput.click();
  });

  fileInput.addEventListener("change", function() {
    if (this.files && this.files.length > 0) {
      handleFilesSelected(this.files);
    }
    this.value = "";
  });

  ["dragenter", "dragover"].forEach(function(event) {
    dropzone.addEventListener(event, function(e) {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.add("drag-over");
    });
  });

  ["dragleave", "drop"].forEach(function(event) {
    dropzone.addEventListener(event, function(e) {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.remove("drag-over");
    });
  });

  dropzone.addEventListener("drop", function(e) {
    var items = e.dataTransfer.items;
    if (items && items.length > 0) {
      handleDroppedItems(items);
    } else if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFilesSelected(e.dataTransfer.files);
    }
  });
}

/* --- Directory drag support (webkitGetAsEntry) --- */

function handleDroppedItems(items) {
  var allFiles = [];

  function processEntry(entry, callback) {
    if (entry.isFile) {
      entry.file(function(file) {
        allFiles.push(file);
        callback();
      }, callback);
    } else if (entry.isDirectory) {
      var reader = entry.createReader();
      var MAX_DEPTH = 10;
      readDirEntries(reader, function() {
        callback();
      });
    } else {
      callback();
    }
  }

  function readDirEntries(reader, callback, depth) {
    depth = depth || 0;
    if (depth >= 10) { callback(); return; }
    reader.readEntries(function(entries) {
      if (entries.length === 0) {
        callback();
        return;
      }
      var pending = entries.length;
      entries.forEach(function(entry) {
        processEntry(entry, function() {
          pending--;
          if (pending === 0) {
            readDirEntries(reader, callback, depth + 1);
          }
        });
      });
    }, callback);
  }

  var fileEntries = [];
  for (var i = 0; i < items.length; i++) {
    var entry = items[i].webkitGetAsEntry ? items[i].webkitGetAsEntry() : null;
    if (entry) {
      fileEntries.push(entry);
    }
  }

  if (fileEntries.length === 0) {
    if (items.length > 0 && items[0].getAsFile) {
      var files = [];
      for (var i = 0; i < items.length; i++) {
        var f = items[i].getAsFile();
        if (f) files.push(f);
      }
      if (files.length > 0) handleFilesSelected(asFileList(files));
    }
    return;
  }

  var pending = fileEntries.length;
  fileEntries.forEach(function(entry) {
    processEntry(entry, function() {
      pending--;
      if (pending === 0) {
        handleFilesSelected(asFileList(allFiles));
      }
    });
  });
}

function asFileList(files) {
  return {
    length: files.length,
    item: function(i) { return files[i]; }
  };
}

/* --- File selection handling --- */

function handleFilesSelected(fileList) {
  var imageFiles = [];
  for (var i = 0; i < fileList.length; i++) {
    var file = fileList[i];
    if (file.type && file.type.startsWith("image/")) {
      imageFiles.push(file);
    }
  }

  if (imageFiles.length === 0) {
    showToast("No image files found", "error");
    return;
  }

  showPreviewItems(imageFiles);
}

function showPreviewItems(files) {
  var list = document.querySelector(".import-preview-list");
  var msg = document.querySelector(".import-dropzone-message");
  if (!list) return;

  msg.style.display = "none";
  list.style.display = "";
  list.innerHTML = "";

  files.forEach(function(file) {
    var item = document.createElement("div");
    item.className = "import-preview-item";

    var img = document.createElement("img");
    var objectUrl = URL.createObjectURL(file);
    img.src = objectUrl;
    item.appendChild(img);

    var name = document.createElement("div");
    name.className = "import-preview-item-name";
    name.textContent = file.name;
    item.appendChild(name);

    item._file = file;
    item._fileName = file.name;
    item._objectUrl = objectUrl;

    list.appendChild(item);
  });
}

/* --- Textarea drag and drop --- */

var textareaDropInitialized = false;

function initTextareaDrop() {
  if (textareaDropInitialized) return;
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;
  textareaDropInitialized = true;

  var dropOverlay = document.createElement("div");
  dropOverlay.className = "textarea-drop-overlay";
  dropOverlay.innerHTML = '<div class="textarea-drop-msg">Drop images to import</div>';
  dropOverlay.style.display = "none";
  textarea.parentNode.appendChild(dropOverlay);

  ["dragenter", "dragover"].forEach(function(event) {
    textarea.addEventListener(event, function(e) {
      if (!e.dataTransfer.types || !Array.from(e.dataTransfer.types).includes("Files")) return;
      e.preventDefault();
      e.stopPropagation();
      dropOverlay.style.display = "";
    });
  });

  ["dragleave", "drop"].forEach(function(event) {
    textarea.addEventListener(event, function(e) {
      dropOverlay.style.display = "none";
    });
  });

  textarea.addEventListener("drop", function(e) {
    e.preventDefault();
    e.stopPropagation();

    var items = e.dataTransfer.items;
    if (!items || items.length === 0) return;

    if (items[0].webkitGetAsEntry) {
      var allFiles = [];
      var pending = items.length;

      for (var i = 0; i < items.length; i++) {
        var entry = items[i].webkitGetAsEntry();
        if (!entry) { pending--; continue; }
        collectFilesFromEntry(entry, allFiles, function() {
          pending--;
          if (pending === 0) {
            processTextareaDropFiles(asFileList(allFiles));
          }
        });
      }
    } else {
      processTextareaDropFiles(e.dataTransfer.files);
    }
  });
}

function collectFilesFromEntry(entry, result, callback) {
  if (entry.isFile) {
    entry.file(function(file) {
      result.push(file);
      callback();
    }, callback);
  } else if (entry.isDirectory) {
    var reader = entry.createReader();
    readDirEntriesFlat(reader, result, callback);
  } else {
    callback();
  }
}

function readDirEntriesFlat(reader, result, callback, depth) {
  depth = depth || 0;
  if (depth >= 10) { callback(); return; }
  reader.readEntries(function(entries) {
    if (entries.length === 0) {
      callback();
      return;
    }
    var pending = entries.length;
    entries.forEach(function(entry) {
      collectFilesFromEntry(entry, result, function() {
        pending--;
        if (pending === 0) {
          readDirEntriesFlat(reader, result, callback, depth + 1);
        }
      });
    });
  }, callback);
}

function processTextareaDropFiles(files) {
  var imageFiles = [];
  for (var i = 0; i < files.length; i++) {
    var file = files.item ? files.item(i) : files[i];
    if (file.type && file.type.startsWith("image/")) {
      imageFiles.push(file);
    }
  }
  if (imageFiles.length === 0) return;

  var dropPos = document.querySelector(".editor-textarea").selectionStart;

  if (imageFiles.length === 1) {
    importSingleFile(imageFiles[0], imageFiles[0].name, function(path) {
      if (path) {
        var textarea = document.querySelector(".editor-textarea");
        textarea.focus();
        textarea.selectionStart = textarea.selectionEnd = dropPos;
        insertTextAtCursor("![](" + encodeURI(path) + ")\n");
      }
    });
  } else {
    importAllFiles(imageFiles.map(function(f) {
      return { file: f, name: f.name };
    }), function() {});
  }
}

/* --- Clipboard paste --- */

var pasteHandlerInitialized = false;

function initPasteHandler() {
  if (pasteHandlerInitialized) return;
  var textarea = document.querySelector(".editor-textarea");
  if (!textarea) return;
  pasteHandlerInitialized = true;

  textarea.addEventListener("paste", function(e) {
    var items = e.clipboardData && e.clipboardData.items;
    if (!items) return;

    var imageItems = [];
    for (var i = 0; i < items.length; i++) {
      if (items[i].type && items[i].type.startsWith("image/")) {
        imageItems.push(items[i]);
      }
    }

    if (imageItems.length === 0) return;

    e.preventDefault();

    if (imageItems.length === 1) {
      var item = imageItems[0];
      var file = item.getAsFile();
      if (!file) return;

      var now = new Date();
      var ts = now.getFullYear() +
        String(now.getMonth() + 1).padStart(2, "0") +
        String(now.getDate()).padStart(2, "0") + "_" +
        String(now.getHours()).padStart(2, "0") +
        String(now.getMinutes()).padStart(2, "0") +
        String(now.getSeconds()).padStart(2, "0");
      var fileName = file.name || "pasted_" + ts + ".png";

      importSingleFile(file, fileName, function(path) {
        if (path) {
          var pos = textarea.selectionStart;
          textarea.focus();
          textarea.selectionStart = textarea.selectionEnd = pos;
          insertTextAtCursor("![](" + encodeURI(path) + ")\n");
        }
      });
    } else {
      var files = [];
      imageItems.forEach(function(item, idx) {
        var f = item.getAsFile();
        if (f) {
          var now = new Date();
          var ts = now.getFullYear() +
            String(now.getMonth() + 1).padStart(2, "0") +
            String(now.getDate()).padStart(2, "0") + "_" +
            String(now.getHours()).padStart(2, "0") +
            String(now.getMinutes()).padStart(2, "0") +
            String(now.getSeconds()).padStart(2, "0");
          var name = f.name || "pasted_" + ts + "_" + idx + ".png";
          files.push({ file: f, name: name });
        }
      });
      if (files.length > 0) {
        importAllFiles(files, function() {});
      }
    }
  });
}

	init();

	document.addEventListener("DOMContentLoaded", function () {
		restoreSidebarState();
	});
})();
