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
      default: "is-medium"
    }
  },
  data() {
    return {
      isFocused: false,
      copied: false
    };
  },
  methods: {
    click() {
      if (!this.isFocused) {
        this.isFocused = true;
        this.select();
      }
    },
    select() {
      let range = document.createRange();
      range.selectNodeContents(this.$refs.textarea);
      let selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
    },
    copy() {
      this.select();
      if (document.execCommand("copy")) {
        this.copied = true;
        window.setTimeout(() => {
          this.copied = false;
        }, 5000); // 5s
      }
    }
  }
};
</script>

<template>
  <div class="has-margin-bottom">
    <div class="field">
      <div class="control">
        <div
          ref="textarea"
          class="textarea"
          rows="bulmaoverride"
          :class="size"
          contenteditable
          @click="click"
          @blur="isFocused = false"
          @focus="select"
          v-text="value"
        ></div>
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
