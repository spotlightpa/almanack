<script>
import "@/css/main.scss";

import PageNav from "./PageNav.vue";

export default {
  components: {
    PageNav
  },
  mounted() {
    this.$auth.$watch("$auth.user", (newUser, oldUser) => {
      if (newUser === oldUser) {
        return;
      }
      let name = newUser ? "home" : "login";
      this.$router.push({ name });
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
