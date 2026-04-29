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

  function setActive(link, links) {
    links.forEach(function (item) {
      item.classList.toggle("active", item === link);
      if (item === link) {
        item.setAttribute("aria-current", "location");
      } else {
        item.removeAttribute("aria-current");
      }
    });
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
