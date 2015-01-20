var defaults = {
  theme: 'default'
}

function loadDefault(key) {
  if (localStorage) {
    try {
      var localDefaults = JSON.parse(localStorage.getItem('defaults'));
      if (localDefaults && localDefaults[key]) {
        return localDefaults[key];
      }
    } catch (e) {
      console.log(e);
      localStorage.removeItem('defaults');
    }
  }
  return defaults[key];
}

function saveDefault(key, value) {
  if (localStorage) {
    try {
      var localDefaults = JSON.parse(localStorage.getItem('defaults'));
      if (!localDefaults) {
        localDefaults = {};
      }
      localDefaults[key] = value;
      localStorage.setItem('defaults', JSON.stringify(localDefaults));
    } catch (e) {
      console.log(e);
      localStorage.removeItem('defaults');
    }
  }
}

function loadCSS(themeName, replace, callback) {
  var link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = '/static/theme/' + themeName + '.css';
  if (callback) {
    link.addEventListener('load', callback, false);
  }
  place = document.getElementsByTagName('link')[0];
  place.parentNode.insertBefore(link, place);
  saveDefault('theme', themeName);
  if (replace) {
    var par = replace.parentNode;
    par.removeChild(replace);
  }
  return link;
}

/* Initialization functions called here. */
window.addEventListener('load', function() {
  var chatName = location.pathname.slice(1);
  chatName = chatName ? chatName : 'General';
  var link = loadCSS(loadDefault('theme'));
  var options = {
    customCommands: [
      [/s(witch)? (.*)/, function(m) {
        window.location.href = m[2];
      }],
      [/t(heme)? (.*)/, function(m) {
        var chat = this;
        link = loadCSS(m[2], link, function(){
          chat.scroll();
        });
      }]
    ]
  };
  var c = new Chat(document.getElementById('chat'), chatName, options);
});


