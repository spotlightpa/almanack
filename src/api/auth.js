import { ref, computed } from "vue";

import netlifyIdentity from "netlify-identity-widget";

const DEV_AUTH_KEY = "almanack_dev_auth";

function loadDevUser() {
  try {
    return JSON.parse(localStorage.getItem(DEV_AUTH_KEY)) ?? null;
  } catch {
    return null;
  }
}

function saveDevUser(user) {
  if (user) {
    localStorage.setItem(DEV_AUTH_KEY, JSON.stringify(user));
  } else {
    localStorage.removeItem(DEV_AUTH_KEY);
  }
}

function makeDevAuth() {
  const user = ref(loadDevUser());

  const isSignedIn = computed(() => !!user.value);
  const roles = computed(() => user.value?.roles ?? []);
  const fullName = computed(() => user.value?.fullName ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  return {
    user,
    isSignedIn,
    roles,
    fullName,
    email,
    isEditor: hasRole("editor"),
    isSpotlightPAUser: hasRole("Spotlight PA"),
    isArcUser: hasRole("arc user"),
    signup() {},
    login() {},
    async logout() {
      user.value = null;
      saveDevUser(null);
    },
    async headers() {
      if (!user.value) return null;
      return { Authorization: "Bearer dev-fake-token" };
    },
    setDevUser({ email, fullName, roles }) {
      user.value = { email, fullName, roles };
      saveDevUser(user.value);
    },
  };
}

function makeAuth() {
  const user = ref(null);

  const token = computed(() => user.value?.token?.access_token ?? null);
  const isSignedIn = computed(() => !!token.value);
  const roles = computed(() => user.value?.app_metadata?.roles ?? []);
  const fullName = computed(() => user.value?.user_metadata?.full_name ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  let methods = {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    async logout() {
      user.value = null;
      try {
        await netlifyIdentity.logout();
      } catch (e) {
        console.warn(e);
        netlifyIdentity.store.user = null;
      }
    },
    async headers() {
      if (!user.value) {
        return null;
      }
      let token;
      try {
        token = await user.value.jwt();
      } catch (e) {
        await methods.logout();
        return null;
      }
      return {
        Authorization: `Bearer ${token}`,
      };
    },
  };

  netlifyIdentity.on("init", async (u) => {
    user.value = u;
    try {
      await u.jwt();
    } catch (err) {
      await methods.logout();
    }
  });
  netlifyIdentity.on("login", (u) => {
    user.value = u;
    netlifyIdentity.close();
  });
  netlifyIdentity.on("logout", () => {
    user.value = null;
  });
  netlifyIdentity.on("error", (err) => {
    console.warn(err);
    user.value = null;
  });

  let APIUrl = window.location.hostname.match(/localhost|\.ts\.net/)
    ? "https://almanack.data.spotlightpa.org/.netlify/identity"
    : null;
  netlifyIdentity.init({ logo: false, APIUrl });

  return {
    user,
    isSignedIn,
    roles,
    fullName,
    email,
    isEditor: hasRole("editor"),
    isSpotlightPAUser: hasRole("Spotlight PA"),
    isArcUser: hasRole("arc user"),
    ...methods,
  };
}

let $auth;

export function useAuth() {
  if (!$auth) {
    $auth = import.meta.env.MODE !== "production" ? makeDevAuth() : makeAuth();
  }
  return $auth;
}
