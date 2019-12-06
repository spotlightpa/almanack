<script>
export default {
  props: {
    article: { type: Object, required: true }
  },
  data() {
    return {
      copied: false,
      viewHTML: false,
      articleHTML: ""
    };
  },
  async mounted() {
    await this.$nextTick();
    this.articleHTML = this.$refs.contentsEls
      .map(comp => comp.$el.outerHTML)
      .join("\n\n");
  },
  methods: {
    async copy(kind) {
      let doHTML = kind === "html";
      if (doHTML != this.viewHTML) {
        this.viewHTML = !this.viewHTML;
        await this.$nextTick();
      }
      if (doHTML) {
        this.selectHTML();
      } else {
        this.selectContent();
      }
      if (document.execCommand("copy")) {
        this.copied = true;
        window.setTimeout(() => {
          this.copied = false;
        }, 5000); // 5s
      }
    },
    selectHTML() {
      this.$refs.htmlEl.select();
    },
    selectContent() {
      let range = document.createRange();
      range.selectNodeContents(this.$refs.richtextEl);
      let selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
    }
  }
};
</script>

<template>
  <div class="block">
    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="buttons has-addons">
            <button
              class="button is-light has-text-weight-semibold"
              type="button"
              @click="viewHTML = false"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'file-word']" />
              </span>
              <span>
                View Rich Text
              </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copy('richtext')"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'copy']" />
              </span>
              <span>
                Copy Rich Text
              </span>
            </button>
          </div>
        </div>
        <div class="level-item">
          <div class="buttons has-addons">
            <button
              class="button is-light has-text-weight-semibold"
              type="button"
              @click="viewHTML = true"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'file-code']" />
              </span>
              <span>
                View HTML
              </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copy('html')"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'copy']" />
              </span>
              <span>
                Copy HTML
              </span>
            </button>
          </div>
        </div>
        <div class="level-item">
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

    <div v-if="!viewHTML" class="message height-50vh">
      <div class="message-body">
        <div
          ref="richtextEl"
          class="content"
          contenteditable
          @focus="selectContent"
        >
          <component
            :is="block.component"
            v-for="(block, i) of article.contentComponents"
            ref="contentsEls"
            :key="i"
            :block="block.block"
          ></component>
        </div>
      </div>
    </div>

    <textarea
      v-if="viewHTML"
      ref="htmlEl"
      class="textarea is-small height-50vh"
      @click.once="selectHTML"
      @focus="selectHTML"
      v-text="articleHTML"
    >
    </textarea>
  </div>
</template>

<style>
.height-50vh {
  height: 50vh;
  overflow-y: scroll;
}
</style>
