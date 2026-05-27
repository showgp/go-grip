(function () {
  var searchInput = null;
  var searchClear = null;
  var noResultsEl = null;
  var searchRaf = null;

  function keepActiveSidebarItemVisible() {
    var sidebar = document.querySelector(".docs-sidebar");
    if (!sidebar) {
      return;
    }

    var active = sidebar.querySelector('a[aria-current="page"], a.active');
    if (!active) {
      return;
    }

    var padding = 24;
    var sidebarRect = sidebar.getBoundingClientRect();
    var activeRect = active.getBoundingClientRect();
    var isAbove = activeRect.top < sidebarRect.top + padding;
    var isBelow = activeRect.bottom > sidebarRect.bottom - padding;
    if (!isAbove && !isBelow) {
      return;
    }

    var targetTop =
      sidebar.scrollTop +
      activeRect.top -
      sidebarRect.top -
      Math.max(padding, (sidebar.clientHeight - active.offsetHeight) / 2);
    sidebar.scrollTop = Math.max(0, targetTop);
  }

  function getLabelText(articleEl) {
    var label = articleEl.querySelector(".docs-sidebar-label");
    return label ? label.textContent || "" : "";
  }

  function findArticles() {
    var sidebar = document.querySelector(".docs-sidebar");
    if (!sidebar) return [];
    var items = sidebar.querySelectorAll(".docs-sidebar-article, .docs-sidebar-directory");
    return Array.prototype.slice.call(items);
  }

  // fuzzyMatch checks whether every character of query appears in text
  // in order, allowing arbitrary gaps between characters (fzf-style matching).
  function fuzzyMatch(query, text) {
    if (query === "") return true;
    var qi = 0;
    for (var ti = 0; ti < text.length && qi < query.length; ti++) {
      if (text[ti] === query[qi]) {
        qi++;
      }
    }
    return qi === query.length;
  }

  function filterArticles(query) {
    var articles = findArticles();
    var lowerQuery = query.toLowerCase().trim();
    var hasMatch = false;

    // First pass: mark each article node
    for (var i = 0; i < articles.length; i++) {
      var el = articles[i];
      var isDir = el.classList.contains("docs-sidebar-directory");
      if (isDir) {
        // Directories are handled in the second pass
        continue;
      }
      var labelText = getLabelText(el).toLowerCase();
      var matches = fuzzyMatch(lowerQuery, labelText);
      if (matches) hasMatch = true;
      if (matches) {
        el.classList.remove("docs-sidebar-article-hidden");
      } else {
        el.classList.add("docs-sidebar-article-hidden");
      }
    }

    // Second pass: update directory visibility based on visible children
    for (var j = 0; j < articles.length; j++) {
      var dirEl = articles[j];
      if (!dirEl.classList.contains("docs-sidebar-directory")) continue;

      var hasVisibleChild = false;
      var childArticles = dirEl.querySelectorAll(".docs-sidebar-article");
      for (var k = 0; k < childArticles.length; k++) {
        if (!childArticles[k].classList.contains("docs-sidebar-article-hidden")) {
          hasVisibleChild = true;
          break;
        }
      }

      if (hasVisibleChild) {
        dirEl.classList.remove("docs-sidebar-directory-hidden");
        var details = dirEl.querySelector(".docs-sidebar-details");
        if (details) details.open = true;
      } else {
        dirEl.classList.add("docs-sidebar-directory-hidden");
      }
    }

    // Show/hide no-results message
    if (noResultsEl) {
      if (!hasMatch && lowerQuery !== "") {
        noResultsEl.classList.remove("docs-sidebar-no-results-hidden");
      } else {
        noResultsEl.classList.add("docs-sidebar-no-results-hidden");
      }
    }

    // Show/hide clear button
    if (searchClear) {
      if (lowerQuery !== "") {
        searchClear.classList.add("visible");
      } else {
        searchClear.classList.remove("visible");
      }
    }
  }

  function onSearchInput() {
    if (searchRaf) {
      window.cancelAnimationFrame(searchRaf);
    }
    searchRaf = window.requestAnimationFrame(function () {
      searchRaf = null;
      filterArticles(searchInput.value);
    });
  }

  function clearSearch() {
    if (searchInput) {
      searchInput.value = "";
      filterArticles("");
      searchInput.focus();
    }
  }

  function setupSearch() {
    searchInput = document.querySelector(".docs-sidebar-search");
    if (!searchInput) return;

    searchClear = document.querySelector(".docs-sidebar-search-clear");
    var nav = document.querySelector(".docs-sidebar nav");

    // Create no-results element
    noResultsEl = document.createElement("p");
    noResultsEl.className = "docs-sidebar-no-results docs-sidebar-no-results-hidden";
    noResultsEl.textContent = "No matching files";
    var sidebar = document.querySelector(".docs-sidebar");
    if (sidebar && nav) {
      sidebar.insertBefore(noResultsEl, nav.nextSibling);
    }

    searchInput.addEventListener("input", onSearchInput);
    searchInput.addEventListener("keydown", function (e) {
      if (e.key === "Escape") {
        clearSearch();
      }
    });

    if (searchClear) {
      searchClear.addEventListener("click", clearSearch);
    }

    // Global Ctrl+F / Cmd+F — focus search box when not in edit mode
    document.addEventListener("keydown", function (e) {
      if ((e.ctrlKey || e.metaKey) && e.key === "f") {
        if (
          document.body.getAttribute("data-editing") === "true" ||
          e.target.isContentEditable ||
          (e.target.tagName &&
            /^(INPUT|SELECT|TEXTAREA)$/i.test(e.target.tagName))
        ) {
          return;
        }
        e.preventDefault();
        searchInput.focus();
        if (searchInput.value) {
          searchInput.select();
        }
      }
    });
  }

  function initSidebar() {
    window.requestAnimationFrame(keepActiveSidebarItemVisible);
    setupSearch();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initSidebar);
  } else {
    initSidebar();
  }
})();
