<template>
  <div>
    <nav>
      <h1>Spotlight PA Almanack</h1>
      <button @click="netlifyOpen('login')">Login</button>
      <button @click="netlifyOpen('signup')">Sign Up</button>
      <button @click="netlifyLogout">Logout</button>
    </nav>
    <main>
      <router-view />
      <h2>User:</h2>
      <pre>
        <code>{{ user|jsonify }}</code>
      </pre>
    </main>
  </div>
</template>

<script>
import netlifyIdentity from "netlify-identity-widget";

netlifyIdentity.init({});

export default {
  data() {
    return {
      user: null
    };
  },
  filters: {
    jsonify(obj) {
      return JSON.stringify(obj, null, 2);
    }
  },
  mounted() {
    ["login", "logout", "signup"].forEach(
      action =>
        void netlifyIdentity.on(action, newUser => {
          console.log(newUser);
          this.user = newUser;
        })
    );
  },
  methods: {
    netlifyOpen(action) {
      netlifyIdentity.open(action);
    },
    netlifyLogout() {
      netlifyIdentity.logout();
    }
  }
};
</script>
