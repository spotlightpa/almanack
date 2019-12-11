// eslint-disable-next-line no-unused-vars
const webpack = require("webpack");

module.exports = {
  devServer: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:3001",
      },
    },
  },
  chainWebpack: config => {
    config.module
      .rule("eslint")
      .use("eslint-loader")
      .options({
        fix: true,
      });
  },
};
