<script>
import "@/css/main.scss";

import PageNav from "./PageNav.vue";

export default {
  components: {
    PageNav
  },
  mounted() {
    this.$auth.$watch("user", (newUser, oldUser) => {
      if (newUser === oldUser) {
        return;
      }
      let route = newUser ? "/home" : "/login";
      this.$router.push(route);
    });
  }
};
</script>

<template>
  <div>
    <PageNav />
    <main>
      <router-view />
    </main>
  </div>
</template>
