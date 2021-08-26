<script>
import { useAuth } from "@/api/hooks.js";

export default {
  name: "ViewLogin",
  metaInfo: {
    title: "Log In",
  },
  setup() {
    let { login, signup, isSignedIn } = useAuth();
    return {
      login,
      signup,
      isSignedIn,
    };
  },
  beforeRouteEnter(to, from, next) {
    next((vm) => {
      if (vm.isSignedIn) {
        next({ name: "home", replace: true });
      }
    });
  },
};
</script>

<template>
  <div class="content">
    <h2>Predicting future sports scores.</h2>
    <p class="has-text-centered">
      <a href="https://en.wikipedia.org/wiki/Benjamin_Franklin">
        <img src="@/assets/img/pr-almanack.jpeg" class="almanack-img" />
      </a>
      <a href="https://en.wikipedia.org/wiki/Back_to_the_Future_Part_II">
        <img src="@/assets/img/sports-alamanc.jpeg" class="almanack-img" />
      </a>
    </p>
    <p>
      Please note that this is an internal content distribution system, not
      intended for public use. Please
      <strong>do not share this URL</strong> with anyone besides the appointed
      contacts at your organization and please be mindful of the notes and
      embargos attached to each story. For assistance or if you have any
      questions, please contact Sarah Anne Hughes (<a
        href="mailto:shughes@spotlightpa.org"
        >shughes@spotlightpa.org</a
      >).
    </p>
    <p>
      To use the <strong>Spotlight PA Almanack</strong>, please
      <a class="has-text-weight-semibold" @click="login">Login</a>
      or
      <a class="has-text-weight-semibold" @click="signup">Sign up</a>.
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
