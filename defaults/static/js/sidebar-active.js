(function () {
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

  function initSidebar() {
    window.requestAnimationFrame(keepActiveSidebarItemVisible);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initSidebar);
  } else {
    initSidebar();
  }
})();
