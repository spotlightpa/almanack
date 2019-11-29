<script>
export default {
  mounted() {
    if (this.$auth.isSignedIn) {
      this.$router.push({ name: "home" });
    }
  }
};
</script>

<template>
  <div class="content">
    <h2>Predicting future sports scores.</h2>
    <p class="has-text-centered">
      <img src="@/assets/img/pr-almanack.jpeg" class="almanack-img" />
      <img src="@/assets/img/sports-alamanc.jpeg" class="almanack-img" />
    </p>
    <p>
      To use the <strong>Spotlight PA Almanack</strong>, please
      <a class="has-text-weight-semibold" @click="$auth.login">Login</a>
      or
      <a class="has-text-weight-semibold" @click="$auth.signup">Sign up</a>.
    </p>
  </div>
</template>

<style scoped>
.almanack-img {
  margin-bottom: 1rem;
  max-height: 40vh;
  max-width: 100%;
  vertical-align: top;
}
</style>
