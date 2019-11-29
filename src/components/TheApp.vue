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
      this.$router.push({ name: "home" });
    });
  }
};
</script>

<template>
  <div>
    <PageNav />
    <main class="section container">
      <router-view />
    </main>
  </div>
</template>
