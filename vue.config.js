const { defineConfig } = require("@vue/cli-service");

module.exports = defineConfig({
  devServer: {
    port: 33159,
    proxy: {
      "/api": {
        target: "http://localhost:33160",
      },
      "/api-background": {
        target: "http://localhost:33160",
      },
      "/ssr": {
        target: "http://localhost:33160",
      },
    },
  },
  chainWebpack: (config) => {
    config.plugin("eslint").tap((options) => {
      options[0].fix = true;
      return options;
    });
  },
  lintOnSave: "warning",
  transpileDependencies: true,
});
