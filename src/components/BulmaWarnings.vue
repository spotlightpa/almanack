<script>
import { useRouter } from "vue-router";

export default {
  props: {
    values: {
      type: Array,
      required: true,
    },
  },
  setup() {
    return { router: useRouter() };
  },
  computed: {
    active() {
      return this.values
        .filter(([cond]) => cond)
        .map(([, link, msg]) => ({ link, msg }));
    },
  },
  methods: {
    async scrollTo(link) {
      let el = document.querySelector(link);
      el.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    },
  },
};
</script>

<template>
  <div v-if="active.length" class="message is-warning">
    <div class="message-header">
      <span>
        <font-awesome-icon :icon="['fas', 'circle-exclamation']" />

        <span
          class="ml-1"
          v-text="active.length === 1 ? 'Warning' : 'Warnings'"
        />
      </span>
    </div>
    <div class="message-body">
      <li v-for="{ link, msg } of active" :key="msg">
        <a @click.prevent="scrollTo(link)">
          {{ msg }}
        </a>
      </li>
    </div>
  </div>
</template>

<style>
html {
  scroll-behavior: smooth;
}
</style>
