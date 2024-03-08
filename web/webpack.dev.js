import {merge} from 'webpack-merge';
import common from "./webpack.common.js"
import webpack from 'webpack'

const config = merge(common, {
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

export default config