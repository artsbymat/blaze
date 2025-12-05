(function() {
  const KEY = "sidebar.details";

  function loadState() {
    try {
      const raw = sessionStorage.getItem(KEY);
      return raw ? JSON.parse(raw) : {};
    } catch {
      return {};
    }
  }

  function saveState(state) {
    try {
      sessionStorage.setItem(KEY, JSON.stringify(state));
    } catch {}
  }

  document.addEventListener("DOMContentLoaded", () => {
    const state = loadState();

    document.querySelectorAll('details[data-key]').forEach(d => {
      const k = d.getAttribute("data-key");

      // restore state
      if (k in state) {
        d.open = state[k];
      }

      // capture change
      d.addEventListener("toggle", () => {
        const s = loadState(); // reload for safety
        s[k] = d.open;
        saveState(s);
      });
    });
  });
})();