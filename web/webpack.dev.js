const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');
const webpack = require('webpack');

module.exports = merge(common, {
    mode: "development",
    devtool: 'inline-source-map',
    plugins: [
        new webpack.EnvironmentPlugin({
            NODE_ENV: "development",
            BASE_URL: "http://localhost:8080"
        }),
    ],
    devServer: {
        static: "./dist",
        headers: {
            "Access-Control-Allow-Origin": "*",
            'Access-Control-Allow-Headers': '*',
            'Access-Control-Allow-Methods': '*',
        },
    },
})
