(function () {
  "use strict";

  var toggle = document.querySelector(".nav-toggle");
  var menu = document.getElementById("nav-menu");

  function setMenuOpen(open) {
    if (!toggle || !menu) return;
    toggle.setAttribute("aria-expanded", String(open));
    toggle.setAttribute("aria-label", open ? "Close menu" : "Open menu");
    menu.classList.toggle("open", open);
    document.body.classList.toggle("nav-open", open);
  }

  if (toggle && menu) {
    toggle.addEventListener("click", function (e) {
      e.stopPropagation();
      var expanded = toggle.getAttribute("aria-expanded") === "true";
      setMenuOpen(!expanded);
    });

    menu.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        setMenuOpen(false);
      });
    });

    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && menu.classList.contains("open")) {
        setMenuOpen(false);
        toggle.focus();
      }
    });

    document.addEventListener("click", function (e) {
      if (!menu.classList.contains("open")) return;
      if (menu.contains(e.target) || toggle.contains(e.target)) return;
      setMenuOpen(false);
    });
  }

  var copyBtn = document.getElementById("copy-btn");
  var installCmd = document.getElementById("install-cmd");
  var copyStatus = document.getElementById("copy-status");

  if (copyBtn && installCmd) {
    copyBtn.addEventListener("click", function () {
      var text = installCmd.textContent.trim();

      function clearCopyStatus() {
        if (!copyStatus) return;
        copyStatus.textContent = "";
        copyStatus.classList.remove("is-error");
      }

      function onSuccess() {
        copyBtn.classList.add("copied");
        copyBtn.setAttribute("aria-label", "Install command copied");
        copyBtn.querySelector(".copy-label").textContent = "Copied!";
        if (copyStatus) {
          copyStatus.classList.remove("is-error");
          copyStatus.textContent = "Copied to clipboard.";
        }
        setTimeout(function () {
          copyBtn.classList.remove("copied");
          copyBtn.setAttribute("aria-label", "Copy install command to clipboard");
          copyBtn.querySelector(".copy-label").textContent = "Copy";
          clearCopyStatus();
        }, 2500);
      }

      function onFail() {
        if (copyStatus) {
          copyStatus.classList.add("is-error");
          copyStatus.textContent =
            "Could not copy automatically. Select the command above and press Ctrl+C.";
        }
      }

      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(onSuccess).catch(function () {
          if (fallbackCopy()) {
            onSuccess();
          } else {
            onFail();
          }
        });
      } else if (fallbackCopy()) {
        onSuccess();
      } else {
        onFail();
      }

      function fallbackCopy() {
        var ta = document.createElement("textarea");
        ta.value = text;
        ta.setAttribute("readonly", "");
        ta.style.position = "fixed";
        ta.style.left = "-9999px";
        document.body.appendChild(ta);
        ta.select();
        var ok = false;
        try {
          ok = document.execCommand("copy");
        } catch (e) {
          ok = false;
        }
        document.body.removeChild(ta);
        return ok;
      }
    });
  }

  var header = document.querySelector(".site-header");
  if (header) {
    window.addEventListener(
      "scroll",
      function () {
        header.style.boxShadow =
          window.scrollY > 10 ? "0 4px 24px rgba(0,0,0,0.3)" : "none";
      },
      { passive: true }
    );
  }
})();
