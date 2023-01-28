<script setup>
const isProd = !!window.location.host.match(/spotlightpa\.org$/);
const isStage = !!window.location.host.match(/netlify\.app$/);
const isDev = !isProd && !isStage;
</script>

<template>
  <div class="body">
    <AppNav id="top-nav" />
    <div>
      <main
        class="section"
        :class="{
          'has-background-success-light': isDev,
          'has-background-warning-light': isStage,
        }"
      >
        <div class="mx-auto content-width">
          <router-view />
        </div>
      </main>
    </div>
    <footer class="footer is-size-4">
      <div class="columns is-centered is-multiline">
        <div class="column is-two-fifths has-text-centered is-size-5">
          <p>
            © <span v-text="new Date().getFullYear()"></span> <br />
            Spotlight PA <br />
            225 Market St Suite 502A <br />
            Harrisburg PA 17101
          </p>
          <p>
            <em>
              All Spotlight PA content is provided “as-is” and its use is at the
              sole risk and discretion of partner news organizations. To view
              the full licensing agreement,
              <a href="/docs/Spotlight PA Content License.docx" download
                >click here</a
              >.
            </em>
          </p>
        </div>
      </div>
    </footer>
  </div>
</template>

<style lang="scss">
@import "@/css/main.scss";
</style>

<style scoped>
.content-width {
  max-width: 80ch;
  width: 100%;
}
</style>
