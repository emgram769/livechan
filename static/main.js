function loadCSS(themeName, replace) {
  if (replace) {
    var par = replace.parentNode;
    par.removeChild(replace);
  }
  var link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = '/static/theme/' + themeName + '.css';
  place = document.getElementsByTagName('link')[0];
  place.parentNode.insertBefore(link, place);
  return link;
}

/* Initialization functions called here. */
window.addEventListener('load', function() {
  var chatName = location.pathname.slice(1);
  chatName = chatName ? chatName : 'General';
  var link = loadCSS('default');
  var options = {
    customCommands: [
      [/s(witch)? (.*)/, function(m) {
        window.location.href = m[2];
      }],
      [/t(heme)? (.*)/, function(m) {
        link = loadCSS(m[2], link);
      }]
    ]
  };
  var c = new Chat(document.getElementById('chat'), chatName, options);
});


