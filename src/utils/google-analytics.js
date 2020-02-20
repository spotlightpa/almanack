// Ensure a Google Analytics window func
if (!window.ga) {
  window.ga = function() {
    (window.ga.q = window.ga.q || []).push(arguments);
  };
  window.ga.l = +new Date();
}

let dnt = !window.location.host.match(/spotlightpa\.org$/);

export function callGA(...args) {
  if (dnt) {
    // eslint-disable-next-line no-console
    console.info("GA", args);
    return;
  }
  window.ga(...args);
}

export function sendGAEvent(ev) {
  callGA("send", "event", ev);
}

export function sendGAPageview(path) {
  callGA("send", "pageview", path);
}

export function setDimensions({ domain, name, role }) {
  callGA("set", "dimension1", domain);
  callGA("set", "dimension2", name);
  callGA("set", "dimension3", role);
}
