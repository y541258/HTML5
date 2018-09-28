self.importScripts('ffmpeg.js');

onmessage = function(e) {
  console.log('ffmpeg_run', ffmpeg_run);
  var files = e.data;
  console.log(files);
  ffmpeg_run({
    arguments: ['-i', '/input/' + files[0].name, '-vf', 'scale=160:90', '-strict', '-2', 'out.mp4'],
    files: files,
  }, function(results) {
    console.log('result',results);
    self.postMessage(results[0].data, [results[0].data]);
  });
}