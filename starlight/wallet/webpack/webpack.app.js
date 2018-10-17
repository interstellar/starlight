// TODO: this should be broken up into `dev` and `prod`
// configuration variants

var webpack = require('webpack')
var getConfig = require('hjs-webpack')
var path = require('path')
var fs = require('fs')

// If you want to make changes to the client and deploy them,
// change this URL to point to where you've deployed your changes.
// See ./sync-frontend.sh for how we're hosting the client.
let publicPath = 'https://starlight-client.s3.amazonaws.com/'
if (process.env.NODE_ENV !== 'production') {
  publicPath = '/'
}

// Set base path to JS and CSS files when
// required by other files
let outPath = 'public'
if (process.env.NODE_ENV !== 'production') {
  outPath = 'node_modules/starlight-react-dlls'
}

// Creates a webpack config object. The
// object can be extended by accessing
// its properties.
var config = getConfig({
  // entry point for the app
  in: 'src/app.tsx',

  // Name or full path of output directory
  // commonly named `www` or `public`. This
  // is where your fully static site should
  // end up for simple deployment.
  out: outPath,

  output: {
    hash: true,
  },

  // This will destroy and re-create your
  // `out` folder before building so you always
  // get a fresh folder. Usually you want this
  // but since it's destructive we make it
  // false by default
  clearBeforeBuild: true,

  html: function(context) {
    return {
      'index.html': context.defaultTemplate({
        publicPath: publicPath,
        head:
          process.env.NODE_ENV !== 'production'
            ? '<script data-dll="true" src="/dependencies.dll.js"></script>'
            : '',
      }),
    }
  },

  port: parseInt(process.env.PORT, 10) || 5000,

  // Proxy API requests to local starlightd server
  devServer: {
    https: {
      key: fs.readFileSync('./starlight/localhost-key.pem'),
      cert: fs.readFileSync('./starlight/localhost.pem'),
    },
    proxy: [
      {
        context: [
          '/api/**',
          '/federation**',
          '/.well-known/**',
          '/starlight/**',
        ],
        options: {
          target: process.env.STARLIGHTD_URL || 'https://localhost:7000',
          secure: false,
        },
      },
    ],
  },
})

// Customize loader configuration
let loaders = config.module.loaders

for (let item of loaders) {
  // Enable babel-loader caching
  if (item.loader == 'babel-loader') {
    item.loader = 'babel-loader?cacheDirectory'
  }
}

config.module.loaders = loaders

// Configure node modules which may or
// may not be present in the browser.
config.node = {
  console: true,
  fs: 'empty',
  net: 'empty',
  tls: 'empty',
}

config.resolve = {
  root: [path.resolve('./src'), path.resolve('./static')],
  extensions: ['', '.js', '.jsx', '.ts', '.tsx'],
}

// module.noParse disables parsing for
// matched files. Used here to bypass
// issues with an AMD configured module.
config.module.noParse = /node_modules\/json-schema\/lib\/validate\.js/

// Import specified env vars in packaged source
const env = {
  'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
  'process.env.SEQCRED': JSON.stringify(process.env.SEQCRED),
  'process.env.PORT': JSON.stringify(process.env.PORT) || '5000',
  'process.env.STARLIGHTD_URL':
    JSON.stringify(process.env.STARLIGHTD_URL) || 'https://localhost:7000',
}

config.plugins.push(new webpack.DefinePlugin(env))

// Enable babel-polyfill
// NOTE: to properly function, 'babel-polyfill' must be the first
// entry loaded. Otherwise, some features will not be present
// to the application runtime.
config.entry.unshift('babel-polyfill')

config.output.publicPath = publicPath

if (process.env.NODE_ENV !== 'production') {
  // Support source maps for Babel
  config.devtool = 'eval-cheap-module-source-map'

  // Use DLL
  config.plugins.push(
    new webpack.DllReferencePlugin({
      context: process.cwd(),
      manifest: require(path.resolve(
        process.cwd(),
        'node_modules/starlight-react-dlls/manifest.json'
      )),
    })
  )
}

module.exports = config
