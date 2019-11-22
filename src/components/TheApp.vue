<script>
export default {
  data() {
    return {
      user: null,
      response: null
    };
  },
  filters: {
    jsonify(obj) {
      return JSON.stringify(obj, null, 2);
    }
  },
  mounted() {
    this.$auth.$watch("user", (newUser, oldUser) => {
      if (newUser === oldUser) {
        return;
      }
      let route = newUser ? "/home" : "/login";
      this.$router.push(route);
    });
  },
  methods: {}
};
</script>

<template>
  <div>
    <nav>
      <h1>Spotlight PA Almanack</h1>
      <button v-if="!$auth.isSignedIn" @click="$auth.login">Login</button>
      <button v-if="!$auth.isSignedIn" @click="$auth.signup">Sign up</button>
      <button v-if="$auth.isSignedIn" @click="$auth.logout">Logout</button>
    </nav>
    <main>
      <router-view />
    </main>
  </div>
</template>
