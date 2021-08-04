// eslint-disable-next-line no-unused-vars
const webpack = require("webpack");

module.exports = {
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
    config.module.rule("eslint").use("eslint-loader").options({
      fix: true,
    });
  },
};
