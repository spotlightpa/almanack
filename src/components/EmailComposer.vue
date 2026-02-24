<script setup>
import { computed, ref } from "vue";
import { post, sendMessage } from "@/api/client-v2.js";

const props = defineProps({
  initialSubject: String,
  initialBody: String,
});

defineEmits(["hide"]);

const rows = Math.max(4, props.initialBody.split("\n").length);

const subject = ref(props.initialSubject);
const body = ref(props.initialBody);
const error = ref(null);
const isSending = ref(false);
const hasSent = ref(false);

const hasChanged = computed(
  () =>
    props.initialSubject !== subject.value || props.initialBody !== body.value
);

async function send() {
  if (window.confirm("Are you sure you want to send this message?")) {
    isSending.value = true;
    [, error.value] = await post(sendMessage, {
      subject: subject.value,
      body: body.value,
    });
    isSending.value = false;
    if (error.value) return;
    hasSent.value = true;
    window.setTimeout(() => {
      hasSent.value = false;
    }, 5000);
  }
}

function discard() {
  if (window.confirm("Are you sure you want to discard your changes?")) {
    subject.value = props.initialSubject;
    body.value = props.initialBody;
  }
}
</script>
<template>
  <div class="box">
    <div class="field">
      <BulmaFieldInput v-model="subject" label="Subject"></BulmaFieldInput>
      <BulmaTextarea v-model="body" label="Body" :rows="rows"></BulmaTextarea>
    </div>
    <ErrorSimple :error="error"></ErrorSimple>
    <div class="buttons">
      <button
        class="button has-text-weight-semibold is-primary"
        :class="{ 'is-loading': isSending }"
        @click="send"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']"></font-awesome-icon>
        </span>
        <span> Send Message </span>
      </button>
      <button
        class="button has-text-weight-semibold is-danger"
        :disabled="isSending || !hasChanged || null"
        @click="discard"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'trash-alt']"></font-awesome-icon>
        </span>
        <span> Discard Changes </span>
      </button>
      <button
        class="button has-text-weight-semibold is-light"
        @click="$emit('hide')"
      >
        Close Composer
      </button>
      <transition name="fade">
        <div
          v-if="hasSent"
          class="tag is-rounded is-success is-light has-text-weight-semibold"
        >
          Message Sent
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.fade {
  transition: all 0.5s ease;
}
</style>
