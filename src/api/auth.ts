import { ref, computed, type Ref } from "vue";
import netlifyIdentity from "netlify-identity-widget";
import type { NetlifyUser } from "netlify-identity-widget";

const DEV_AUTH_KEY = "almanack_dev_auth";

export interface AuthUser {
  email: string;
  fullName: string;
  roles: string[];
}

export interface Auth {
  user: Readonly<Ref<AuthUser | null>>;
  isSignedIn: Readonly<Ref<boolean>>;
  roles: Readonly<Ref<string[]>>;
  fullName: Readonly<Ref<string>>;
  email: Readonly<Ref<string>>;
  isEditor: Readonly<Ref<boolean>>;
  isSpotlightPAUser: Readonly<Ref<boolean>>;
  isArcUser: Readonly<Ref<boolean>>;
  signup(): void;
  login(): void;
  logout(): Promise<void>;
  headers(): Promise<Record<string, string> | null>;
  setUser(u: AuthUser): void;
}

function loadDevUser(): AuthUser | null {
  try {
    return (
      (JSON.parse(
        localStorage.getItem(DEV_AUTH_KEY) ?? "null"
      ) as AuthUser | null) ?? null
    );
  } catch {
    return null;
  }
}

function saveDevUser(user: AuthUser | null) {
  if (user) {
    localStorage.setItem(DEV_AUTH_KEY, JSON.stringify(user));
  } else {
    localStorage.removeItem(DEV_AUTH_KEY);
  }
}

function makeDevAuth(): Auth {
  const user = ref<AuthUser | null>(loadDevUser());

  const isSignedIn = computed(() => !!user.value);
  const roles = computed(() => user.value?.roles ?? []);
  const fullName = computed(() => user.value?.fullName ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name: string) {
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
    async headers(): Promise<Record<string, string> | null> {
      if (!user.value) return null;
      return { Authorization: "Bearer dev-fake-token" };
    },
    setUser(u: AuthUser) {
      user.value = u;
      saveDevUser(u);
    },
  };
}

function makeAuth(): Auth {
  const netlifyUser = ref<NetlifyUser | null>(null);

  function toAuthUser(u: NetlifyUser): AuthUser {
    return {
      email: u.email,
      fullName: u.user_metadata?.full_name ?? "",
      roles: u.app_metadata?.roles ?? [],
    };
  }
  const user = ref<AuthUser | null>(null);
  const isSignedIn = computed(() => !!user.value);
  const roles = computed(() => user.value?.roles ?? []);
  const fullName = computed(() => user.value?.fullName ?? "");
  const email = computed(() => user.value?.email ?? "");

  function hasRole(name: string) {
    return computed(() =>
      roles.value.some((role) => role === name || role === "admin")
    );
  }

  const methods = {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    async logout() {
      netlifyUser.value = null;
      try {
        await netlifyIdentity.logout();
      } catch (e) {
        console.warn(e);
        netlifyIdentity.store.user = null;
      }
    },
    async headers(): Promise<Record<string, string> | null> {
      if (!netlifyUser.value) {
        return null;
      }
      let token: string;
      try {
        token = await netlifyUser.value.jwt();
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
    netlifyUser.value = u;
    user.value = u ? toAuthUser(u) : null;
    try {
      await u?.jwt();
    } catch {
      await methods.logout();
    }
  });
  netlifyIdentity.on("login", (u) => {
    netlifyUser.value = u;
    user.value = toAuthUser(u);
    netlifyIdentity.close();
  });
  netlifyIdentity.on("logout", () => {
    netlifyUser.value = null;
    user.value = null;
  });
  netlifyIdentity.on("error", (err) => {
    console.warn(err);
    netlifyUser.value = null;
    user.value = null;
  });

  const APIUrl = window.location.hostname.match(/localhost|\.ts\.net/)
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
    setUser(u: AuthUser) {
      user.value = u;
    },
    ...methods,
  };
}

let $auth: Auth | null = null;

export function useAuth() {
  if (!$auth) {
    $auth = import.meta.env.MODE !== "production" ? makeDevAuth() : makeAuth();
  }
  return $auth;
}
