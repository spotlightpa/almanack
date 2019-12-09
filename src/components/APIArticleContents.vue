<script>
import DOMInnerHTML from "./DOMInnerHTML.vue";

export default {
  components: {
    DOMInnerHTML
  },
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
  computed: {
    embeds() {
      return this.article.embedComponents;
    }
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
    <h2 v-if="embeds.length === 1" class="title">
      Embed
    </h2>
    <h2 v-if="embeds.length > 1" class="title">Embeds: {{ embeds.length }}</h2>

    <component
      :is="component"
      v-for="{ block, component, n } of embeds"
      :key="n"
      :block="block"
      :n="n"
    ></component>

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

    <div
      v-if="!viewHTML"
      class="textarea height-50vh"
      contenteditable
      @focus="selectContent"
    >
      <div ref="richtextEl" class="content">
        <component
          :is="block.component"
          v-for="(block, i) of article.contentComponents"
          ref="contentsEls"
          :key="i"
          :block="block.block"
        ></component>
      </div>
    </div>

    <DOMInnerHTML @mounted="articleHTML = $event">
      <component
        :is="block.component"
        v-for="(block, i) of article.htmlComponents"
        :key="i"
        :block="block.block"
      ></component>
    </DOMInnerHTML>

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
