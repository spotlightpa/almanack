<script>
export default {
  name: "BulmaCopyInput",
  props: {
    value: String,
    rows: { type: Number, default: 1 },
    label: {
      type: String,
      default: "text"
    },
    size: {
      type: String,
      default: "is-large"
    }
  },
  data() {
    return {
      copied: false
    };
  },
  methods: {
    copy() {
      this.$refs.textarea.select();
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
  <div class="has-margin-bottom">
    <div class="field">
      <div class="control">
        <textarea
          ref="textarea"
          class="textarea"
          :class="size"
          :rows="rows"
          readonly
          @click.once="select"
          @focus="select"
          v-text="value"
        >
        </textarea>
      </div>
    </div>
    <div class="field">
      <div class="buttons">
        <button
          type="button"
          class="button is-primary has-text-weight-semibold"
          title="Copy"
          @click="copy"
        >
          <span class="icon">
            <font-awesome-icon :icon="['far', 'copy']" />
          </span>
          <span> Copy {{ label }} </span>
        </button>
        <transition name="fade">
          <div
            v-if="copied"
            class="tag is-rounded is-success is-light has-text-weight-semibold"
          >
            Copied
          </div>
        </transition>
      </div>
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
