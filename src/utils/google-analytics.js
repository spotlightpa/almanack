// Ensure a gtag stub exists so calls made before gtag.js loads still queue
// up via the dataLayer.
if (!window.gtag) {
  window.dataLayer = window.dataLayer || [];
  window.gtag = function () {
    window.dataLayer.push(arguments);
  };
}

let dnt = !window.location.host.match(/spotlightpa\.org$/);

export function callGA(...args) {
  if (dnt) {
    console.info("GA", args);
    return;
  }
  window.gtag(...args);
}

// `eventAction` becomes the GA4 event name; the rest become event parameters.
export function sendGAEvent({
  eventCategory,
  eventAction,
  eventLabel,
  eventValue,
  ...rest
} = {}) {
  let params = { ...rest };
  if (eventCategory !== undefined) params.event_category = eventCategory;
  if (eventLabel !== undefined) params.event_label = eventLabel;
  if (eventValue !== undefined) params.value = eventValue;
  callGA("event", eventAction, params);
}

export function sendGAPageview(path) {
  callGA("event", "page_view", { page_path: path });
}

// Register `user_domain`, `user_name`, and `user_role` as user-scoped custom
// dimensions in the GA4 admin to surface them in reports.
export function setDimensions({ domain, name, role }) {
  callGA("set", "user_properties", {
    user_domain: domain,
    user_name: name,
    user_role: role,
  });
}
