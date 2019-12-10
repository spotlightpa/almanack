<script>
import "@/css/main.scss";

import PageNav from "./PageNav.vue";

export default {
  components: {
    PageNav,
  },
  mounted() {
    this.$auth.$watch("$auth.user", (newUser, oldUser) => {
      if (newUser === oldUser) {
        return;
      }
      let name = newUser ? "articles" : "login";
      this.$router.push({ name });
    });
  },
};
</script>

<template>
  <div>
    <PageNav id="top-nav" />
    <main class="section container">
      <div class="columns is-centered">
        <div class="column is-four-fifths">
          <router-view />
        </div>
      </div>
    </main>
  </div>
</template>
