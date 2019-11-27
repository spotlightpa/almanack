<script>
import netlifyIdentity from "netlify-identity-widget";

export default {
  data() {
    return {
      user: null,
      error: null
    };
  },
  computed: {
    token() {
      return (
        (this.user && this.user.token && this.user.token.access_token) || null
      );
    },
    isSignedIn() {
      return !!this.token;
    },
    roles() {
      return (
        (this.isSignedIn &&
          this.user.app_metadata &&
          this.user.app_metadata.roles) ||
        []
      );
    }
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
  methods: {
    signup() {
      netlifyIdentity.open("signup");
    },
    login() {
      netlifyIdentity.open("login");
    },
    logout() {
      netlifyIdentity.logout();
    },
    async headers() {
      let token;
      try {
        token = await this.user.jwt(true);
      } catch (e) {
        this.logout();
        throw e;
      }
      return {
        Authorization: `Bearer ${token}`
      };
    },
    hasRole(name) {
      return this.roles.some(role => role === name || role === "admin");
    }
  }
};
</script>
