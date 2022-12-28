<script setup>
import { computed, ref } from "vue";

import { intcomma } from "journalize";

import { formatDate } from "@/utils/time-format.js";
import BulmaDateTime from "./BulmaDateTime.vue";

const isOpen = ref(false);
const showComposer = ref(false);
const isDirty = ref(false);
const article = ref(null);
const status = ref(null);
const note = ref("");
const embargo = ref(null);

function open(obj) {
  isOpen.value = true;
  showComposer.value = false;
  isDirty.value = false;
  article.value = obj;
  status.value = obj.status;
  note.value = obj.note;
  embargo.value = obj.embargoUntil;
}

function close() {
  if (!isDirty.value) {
    isOpen.value = false;
    article.value = null;
  } else if (window.confirm("Discard unsaved changes?")) {
    isOpen.value = false;
    article.value = null;
  }
}

function statusClass(val) {
  return {
    "is-white": status.value !== val,
    "is-primary": status.value === val,
  };
}

const emailBody = computed(() => {
  let a = article.value;
  let noteText = note.value ? `\n\nPublication Notes:\n\n${note.value}` : "";
  let text = `
New ${a.slug}

https://almanack.data.spotlightpa.org/shared-articles/${a.id}

Planned for ${formatDate(a.arc.plannedDate)}${noteText}

Budget:

${a.arc.budgetLine}

Word count planned: ${intcomma(a.arc.plannedWordCount)}
Word count actual: ${intcomma(a.arc.actualWordCount)}
Lines: ${a.arc.actualLineCount}
Column inches: ${a.arc.actualInchCount}
`;
  return text.trim();
});

defineExpose({ open });
</script>

<template>
  <BulmaModal :model-value="isOpen" @update:modelValue="close">
    <article v-if="article" class="message is-primary">
      <div class="message-header">
        <p>
          <font-awesome-icon :icon="['far', 'newspaper']" /> {{ article.slug }}
        </p>
        <button class="delete" aria-label="close" @click="close"></button>
      </div>
      <div class="message-body">
        <div v-if="!article.pageID" class="mb-5">
          <div class="label">Import to Spotlight PA</div>
          <div class="buttons">
            <button class="button is-primary has-text-weight-semibold">
              As News article
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              @click="close"
            >
              As State College article
            </button>
          </div>
        </div>

        <h3 class="label">Status</h3>
        <div class="buttons">
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('imported')"
            type="button"
            @click="
              status = 'imported';
              isDirty = true;
            "
          >
            Imported
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('preview')"
            type="button"
            @click="
              status = 'preview';
              isDirty = true;
            "
          >
            Preview
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('embargo')"
            type="button"
            @click="
              status = 'embargo';
              isDirty = true;
            "
          >
            Embargo
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('released')"
            type="button"
            @click="
              status = 'released';
              isDirty = true;
            "
          >
            Released
          </button>
        </div>

        <BulmaDateTime
          v-if="status === 'embargo'"
          :model-value="null"
          label="Embargo time"
          help="List the latest time that an article will be under embargo for partners."
        />

        <BulmaTextarea
          :model-value="note"
          label="Note"
          help="Additional clarifications and instructions for partners"
          @update:modelValue="
            isDirty = true;
            note = $event;
          "
        />

        <button class="button is-warning has-text-weight-semibold">
          Refresh from Arc
        </button>

        <div class="mt-5 buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="!isDirty || null"
          >
            Save changes
          </button>
          <button
            class="button is-danger has-text-weight-semibold"
            :disabled="!isDirty || null"
            @click="close"
          >
            Discard changes
          </button>
        </div>
        <button
          type="button"
          class="button is-small has-text-weight-semibold"
          :class="showComposer ? 'is-danger' : 'is-primary'"
          @click="showComposer = !showComposer"
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'paper-plane']" />
          </span>
          <span
            v-text="!showComposer ? 'Compose Message' : 'Discard Message'"
          />
        </button>

        <EmailComposer
          v-if="showComposer"
          class="mt-5"
          :initial-subject="`New Spotlight PA story ${article.slug}`"
          :initial-body="emailBody"
          @hide="showComposer = false"
        />
      </div>
    </article>
  </BulmaModal>
</template>
