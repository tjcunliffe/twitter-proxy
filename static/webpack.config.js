// var CommonsChunkPlugin = require("./node_modules/webpack/lib/optimize/CommonsChunkPlugin");

module.exports = {
    entry: {
        statechange: "./scripts/statechange.jsx"
    },
    output: {
        path: "./src",
        filename: "[name]-bundle.js"
    },
    module: {
      loaders: [
          {
              //regex for file type supported by the loader
              test: /\.(jsx)$/,

              //type of loader to be used
              //loaders can accept parameters as a query string
              loader: 'babel'
          },
          {
              test: /\.js$/, loader: 'babel-loader'
          }
      ]
    },
    plugins: [
        //new CommonsChunkPlugin("commons.js")
    ]
};

