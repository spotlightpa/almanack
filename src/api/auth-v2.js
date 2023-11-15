import { computed, ref, effect } from "vue";

import GoTrue from "gotrue-js";

const auth = new GoTrue({
  APIUrl: "https://almanack.data.spotlightpa.org/.netlify/identity",
  setCookie: true,
});

const lastModified = ref(1);
const loginError = ref(null);

effect(() => {
  if (loginError.value) console.log("loginError", loginError.value);
});

function init() {
  let hash = window.location.hash.slice(1); // remove # at start
  if (/access_token/.test(hash)) {
    console.log("have hash");
    let searchParams = new URLSearchParams(hash);
    let params = {
      access_token: searchParams.get("access_token"),
      expires_in: searchParams.get("expires_in"),
      refresh_token: searchParams.get("refresh_token"),
      token_type: searchParams.get("token_type"),
    };
    auth
      .createUser(params, true)
      .then(() => {
        loginError.value = null;
        lastModified.value++;
      })
      .catch((err) => {
        loginError.value = err;
        lastModified.value++;
      });
  }
}

export const fullName = computed(
  () =>
    lastModified.value && (auth.currentUser()?.user_metadata?.full_name ?? "")
);

export async function signup() {}

export async function login(email, password) {
  try {
    await auth.login(email, password, true);
    loginError.value = null;
  } catch (err) {
    loginError.value = err;
  }
  lastModified.value++;
}

export async function logout() {
  await auth.currentUser().logout();
  lastModified.value++;
}

init();
