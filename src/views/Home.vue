<template>
  <div>
    <hero />
    <hr />
    <home-content />
    <div id="modal" @click="openModal">
      modal
    </div>
  </div>
</template>

<script>
const netlifyIdentity = require("netlify-identity-widget");

import Hero from "../components/Hero";
import HomeContent from "../components/HomeContent";

export default {
  name: "home",
  components: {
    Hero,
    HomeContent
  },
  mounted() {
    netlifyIdentity.init({
      container: "#modal"
    });
  },
  methods: {
    openModal() {
      netlifyIdentity.open();
    }
  }
};
</script>

<style lang="scss" scoped>
.next-steps {
  .fa-link {
    margin-right: 5px;
  }
}
</style>
