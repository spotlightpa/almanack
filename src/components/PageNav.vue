<script>
export default {
  data() {
    return {
      isOpen: false
    };
  },
  methods: {
    menuToggle() {
      this.isOpen = !this.isOpen;
    }
  }
};
</script>

<template>
  <nav class="navbar" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <router-link
        to="/"
        exact
        class="navbar-item is-size-3 has-text-white has-text-weight-bold"
      >
        <img src="@/assets/circle-white-on-trans.svg" alt="Spotlight PA logo" />
        Almanack
      </router-link>

      <a
        role="button"
        class="navbar-burger burger"
        aria-label="menu"
        aria-expanded="false"
        @click.prevent="menuToggle"
      >
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
        <span aria-hidden="true"></span>
      </a>
    </div>

    <div class="navbar-menu" :class="{ 'is-active': isOpen }">
      <div class="navbar-end">
        <div class="navbar-item">
          <div class="buttons">
            <button
              v-if="!$auth.isSignedIn"
              class="button is-primary has-text-weight-semibold"
              @click="$auth.login"
            >
              Login
            </button>
            <button
              v-if="!$auth.isSignedIn"
              class="button is-success has-text-weight-semibold"
              @click="$auth.signup"
            >
              Sign up
            </button>
            <button
              v-if="$auth.isSignedIn"
              class="button is-warning has-text-weight-semibold"
              @click="$auth.logout"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </div>
  </nav>
</template>
