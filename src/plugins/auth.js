import Vue from "vue";
import netlifyIdentity from "netlify-identity-widget";

let authComponent = new Vue({
  data() {
    return {
      user: null,
      error: null
    };
  },
  created() {
    netlifyIdentity.on("init", user => {
      this.user = user;
    });
    netlifyIdentity.on("login", user => {
      this.user = user;
      netlifyIdentity.close();
    });
    netlifyIdentity.on("logout", () => {
      this.user = null;
    });
    netlifyIdentity.on("error", err => {
      this.error = err;
    });
    netlifyIdentity.init({ logo: false });
  },
  computed: {
    token() {
      return (
        (this.user && this.user.token && this.user.token.access_token) || null
      );
    },
    isSignedIn() {
      return !!this.token;
    }
  },
  methods: {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    logout() {
      netlifyIdentity.logout();
    }
  }
});

export let AuthPlugin = {
  install(Vue) {
    Vue.prototype.$auth = authComponent;
  }
};

export function authGuard(to, from, next) {
  authComponent.isSignedIn ? next() : next("login");
}
