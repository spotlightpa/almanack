<script>
export default {
  name: "BulmaCopyInput",
  props: { value: String },
  data() {
    return {
      copied: false
    };
  },
  methods: {
    copy() {
      this.$refs.input.select();
      if (document.execCommand("copy")) {
        this.copied = true;
        window.setTimeout(() => {
          this.copied = false;
        }, 5000); // 5s
      }
    },
    select(ev) {
      ev.target.select();
    }
  }
};
</script>

<template>
  <div>
    <div class="field has-addons">
      <div class="control">
        <button
          type="button"
          class="button is-primary"
          title="Copy"
          @click="copy"
        >
          <font-awesome-icon :icon="['far', 'copy']" />
        </button>
      </div>
      <div class="control is-expanded">
        <input
          ref="input"
          class="input is-fullwidth"
          type="text"
          :value="value"
          readonly
          @focus="select"
        />
      </div>
    </div>

    <div class="help is-primary fade" :class="{ 'is-invisible': !copied }">
      Copied
    </div>
  </div>
</template>

<style scoped>
.fade {
  transition: all 0.5s ease;
}

.is-invisible {
  opacity: 0;
}
</style>
