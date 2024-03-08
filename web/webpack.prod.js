import { merge } from 'webpack-merge'
import common from './webpack.common.js'
import webpack from 'webpack'

const config = merge(common, {
    mode: "production",
    plugins: [
        new webpack.EnvironmentPlugin({
            NODE_ENV: "production",
            BASE_URL: "",
        })
    ],
    devtool: 'source-map',
})

export default config
