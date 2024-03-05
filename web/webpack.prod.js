const { merge } = require('webpack-merge');
const common = require('./webpack.common.js');

module.exports = merge(common, {
    mode: "production",
    plugins: [
        new webpack.EnvironmentPlugin({
            NODE_ENV: "production",
            BASE_URL: "",
        });
    ],
    devtool: 'source-map',
})
