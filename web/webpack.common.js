import path from 'path';
import {fileURLToPath} from 'url';
import HtmlWebpackPlugin from 'html-webpack-plugin';
import CopyPlugin from 'copy-webpack-plugin';
import DotenvWebpackPlugin from 'dotenv-webpack';

const config = {
    entry: './src/app.ts',
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
        ],
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: 'src/index.html',
            filename: 'index.html',
        }),
        new DotenvWebpackPlugin(),
        new CopyPlugin({
            patterns: [
                { from: "**/*", to: "openmct/", context: "node_modules/openmct/dist"},
            ]
        })
    ],
    resolve: {
        extensions: ['.tsx', '.ts', '.js'],
    },
    externals: {
        openmct: "openmct",
    },
    output: {
        filename: 'main.js',
        path: path.resolve(path.dirname(fileURLToPath(import.meta.url)), 'dist'),
    },
};

export default config