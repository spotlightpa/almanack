<script>
export default {
  data() {
    return { userInfo: null, error: null };
  },
  methods: {
    async getUserInfo() {
      let headers = await this.$auth.headers();
      let [data, err] = await fetch("/api/user-info", {
        headers
      })
        .then(resp => resp.json())
        .then(data => [data, null])
        .catch(err => [null, err]);
      this.userInfo = data;
      this.error = err;
    }
  }
};
</script>

<template>
  <div class="section container content">
    <h2>
      Hello {{ $auth.user.user_metadata.full_name }} (<span
        v-for="role of $auth.user.app_metadata.roles"
        :key="role"
        v-text="role"
      ></span
      >).
    </h2>
    <button
      class="button is-primary has-text-weight-semibold"
      type="button"
      @click="getUserInfo"
    >
      Get User Info
    </button>
    <p>
      {{ userInfo }}
    </p>
    <p v-if="error" class="message is-danger" v-text="error"></p>
  </div>
</template>
