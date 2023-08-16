const path = require('path');

module.exports = {
  mode: 'production',
  entry: {
    simple: './simple.js',
  },
  output: {
    path: path.resolve(__dirname, 'dist'), // eslint-disable-line
    libraryTarget: 'commonjs',
    filename: '[name].bundle.js',
  },
  module: {
    rules: [{ test: /\.js$/, use: 'babel-loader' }],
  },
  target: ['web', 'es5'],
  externals: /k6(\/.*)?/,
};
