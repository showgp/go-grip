(function () {
  function isEditableTarget(target) {
    if (!target || target === document.body) {
      return false;
    }

    var tagName = target.tagName ? target.tagName.toLowerCase() : "";
    return (
      target.isContentEditable ||
      tagName === "input" ||
      tagName === "select" ||
      tagName === "textarea"
    );
  }

  function hasTextSelection() {
    var selection = window.getSelection ? window.getSelection() : null;
    return selection && selection.type === "Range";
  }

  function navigateTo(path) {
    if (!path) {
      return;
    }
    window.location.href = path;
  }

  function initArticleNav() {
    var previousPath = document.body.getAttribute("data-prev-article");
    var nextPath = document.body.getAttribute("data-next-article");
    if (!previousPath && !nextPath) {
      return;
    }

    document.addEventListener("keydown", function (event) {
      if (
        event.defaultPrevented ||
        event.altKey ||
        event.ctrlKey ||
        event.metaKey ||
        event.shiftKey ||
        document.body.getAttribute("data-editing") === "true" ||
        isEditableTarget(event.target) ||
        hasTextSelection()
      ) {
        return;
      }

      if (event.key === "ArrowLeft" && previousPath) {
        event.preventDefault();
        navigateTo(previousPath);
      } else if (event.key === "ArrowRight" && nextPath) {
        event.preventDefault();
        navigateTo(nextPath);
      }
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initArticleNav);
  } else {
    initArticleNav();
  }
})();
