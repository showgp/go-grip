(function () {
  function cssEscape(value) {
    if (window.CSS && typeof window.CSS.escape === "function") {
      return window.CSS.escape(value);
    }
    return value.replace(/["\\]/g, "\\$&");
  }

  function findTarget(link) {
    var href = link.getAttribute("href");
    if (!href || href.charAt(0) !== "#") {
      return null;
    }

    try {
      var id = decodeURIComponent(href.slice(1));
      return document.getElementById(id) || document.querySelector('[name="' + cssEscape(id) + '"]');
    } catch (err) {
      return null;
    }
  }

  function keepActiveLinkVisible(link) {
    var toc = link.closest(".docs-toc");
    if (!toc) {
      return;
    }

    var padding = 24;
    var linkTop = link.offsetTop;
    var linkBottom = linkTop + link.offsetHeight;
    var visibleTop = toc.scrollTop;
    var visibleBottom = visibleTop + toc.clientHeight;

    if (linkTop < visibleTop + padding) {
      toc.scrollTop = Math.max(0, linkTop - padding);
    } else if (linkBottom > visibleBottom - padding) {
      toc.scrollTop = linkBottom - toc.clientHeight + padding;
    }
  }

  function setActive(link, links) {
    links.forEach(function (item) {
      item.classList.toggle("active", item === link);
      if (item === link) {
        item.setAttribute("aria-current", "location");
      } else {
        item.removeAttribute("aria-current");
      }
    });

    if (link) {
      keepActiveLinkVisible(link);
    }
  }

  function scrollBehavior() {
    if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      return "auto";
    }
    return "smooth";
  }

  function scrollY() {
    return window.scrollY || window.pageYOffset || document.documentElement.scrollTop || 0;
  }

  function documentHeight() {
    var body = document.body;
    var doc = document.documentElement;
    return Math.max(
      body.scrollHeight,
      body.offsetHeight,
      doc.clientHeight,
      doc.scrollHeight,
      doc.offsetHeight
    );
  }

  function targetTop(target) {
    return target.getBoundingClientRect().top + scrollY();
  }

  function initTOC() {
    var toc = document.querySelector(".docs-toc");
    if (!toc) {
      return;
    }

    var links = Array.prototype.slice.call(toc.querySelectorAll('a[href^="#"]'));
    if (links.length === 0) {
      return;
    }

    var pairs = links
      .map(function (link) {
        return {
          link: link,
          target: findTarget(link),
        };
      })
      .filter(function (pair) {
        return pair.target;
      });

    if (pairs.length === 0) {
      return;
    }

    function activePairForScroll() {
      if (scrollY() + window.innerHeight >= documentHeight() - 2) {
        return pairs[pairs.length - 1];
      }

      var marker = scrollY() + 96;
      var activePair = pairs[0];

      for (var i = 0; i < pairs.length; i += 1) {
        if (targetTop(pairs[i].target) <= marker) {
          activePair = pairs[i];
        } else {
          break;
        }
      }

      return activePair;
    }

    var scheduled = false;
    function updateActive() {
      scheduled = false;
      setActive(activePairForScroll().link, links);
    }

    function scheduleUpdate() {
      if (scheduled) {
        return;
      }
      scheduled = true;
      window.requestAnimationFrame(updateActive);
    }

    links.forEach(function (link) {
      link.addEventListener("click", function (event) {
        var target = findTarget(link);
        if (!target) {
          return;
        }

        event.preventDefault();
        setActive(link, links);
        history.pushState(null, "", link.getAttribute("href"));
        target.scrollIntoView({
          behavior: scrollBehavior(),
          block: "start",
        });
        window.setTimeout(scheduleUpdate, 80);
      });
    });

    window.addEventListener("scroll", scheduleUpdate, { passive: true });
    window.addEventListener("resize", scheduleUpdate);
    window.addEventListener("hashchange", scheduleUpdate);

    var hashLink = links.find(function (link) {
      return link.getAttribute("href") === window.location.hash;
    });
    setActive(hashLink || pairs[0].link, links);
    scheduleUpdate();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initTOC);
  } else {
    initTOC();
  }
})();
