<script>
import "@/css/main.scss";

import PageNav from "./PageNav.vue";

export default {
  components: {
    PageNav,
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
