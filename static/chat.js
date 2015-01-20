/* @file chat.js
 * 
 * Include this file to spawn a livechan chat.
 * Use Chat(domElement, channel).
 */

/* @brief Creates a structure of html elements for the
 *        chat.
 *
 * @param domElem The element to be populated with the
 *        chat structure.
 * @return An object of references to the structure
 *         created.
 */
function buildChat(domElem) {
  var output = document.createElement('div');
  output.className = 'livechan_chat_output';

  var input = document.createElement('form');
  input.className = 'livechan_chat_input';
  
  var name = document.createElement('input');
  name.className = 'livechan_chat_input_name';
  name.setAttribute('placeholder', 'Anonymous');

  var file = document.createElement('input');
  file.className = 'livechan_chat_input_file';
  file.setAttribute('type', 'button');
  file.setAttribute('value', 'upload');
  file.setAttribute('onclick', 'alert("not yet implemented");');

  var messageDiv = document.createElement('div');
  messageDiv.className = 'livechan_chat_input_message_div';
  
  var message = document.createElement('textarea');
  message.className = 'livechan_chat_input_message';

  var submit = document.createElement('input');
  submit.className = 'livechan_chat_input_submit';
  submit.setAttribute('type', 'submit');
  submit.setAttribute('value', 'send');

  input.appendChild(name);
  input.appendChild(file);
  messageDiv.appendChild(message);
  input.appendChild(messageDiv);
  input.appendChild(submit);

  domElem.appendChild(output);
  domElem.appendChild(input);

  return {
    output: output,
    input: {
      form: input,
      message: message,
      name: name,
      submit: submit,
      file: file
    }
  };
}

function Connection(ws, channel) {
  this.ws = ws;
  this.channel = channel;
}

Connection.prototype.send = function(obj) {
  /* Jsonify the object and send as string. */
  if (this.ws) {
    this.ws.send(JSON.stringify(obj));
  }
}

Connection.prototype.onmessage = function(callback) {
  this.ws.onmessage = function(event) {
    try {
      var data = JSON.parse(event.data);
      callback(data);
    } catch (e) {
      /* Ignore the error. */
    }
  }
}

Connection.prototype.onclose = function(callback) {
  this.ws.onclose = callback;
}

/* @brief Initializes the websocket connection.
 *
 * @param channel The channel to open a connection to.
 * @return A connection the the websocket.
 */
function initWebSocket(channel, connection) {
  var ws = null;
  if (window['WebSocket']) {
    try {
      ws = new WebSocket('ws://'+location.host+'/ws/'+channel);
    } catch(e) {
      ws = null;
    }
  }
  if (ws !== null) {
    ws.onerror = function() {
      connection.ws = null;
    };
    if (connection) {
      console.log("reconnected.");
      connection.ws = ws;
      return connection;
    } else {
      return new Connection(ws, channel);
    }
  } else {
    return null;
  }
}

/* @brief Parses and returns a message div.
 *
 * @param data The message data to be parsed.
 * @return A dom element containing the message.
 */
function parse(text, rules, end_tag) {
  var output = document.createElement('div'); 
  var position = 0;
  var end_matched = false;
  if (end_tag) {
    var end_handler = function(m) {
      end_matched = true;
    }
    rules = [[end_tag, end_handler]].concat(rules);
  }
  do {
    var match = null;
    var match_pos = text.length;
    var handler = null;
    for (var i = 0; i < rules.length; i++) {
      rules[i][0].lastIndex = position;
      var result = rules[i][0].exec(text);
      if (result !== null && position <= result.index && result.index < match_pos) {
        match = result;
        match_pos = result.index;
        handler = rules[i][1];
      }
    }
    var unmatched_text = text.substring(position, match_pos);
    output.appendChild(document.createTextNode(unmatched_text));
    position = match_pos;
    if (match !== null) {
      position += match[0].length;
      output.appendChild(handler(match));
    }
  } while (match !== null && !end_matched);
  return output;
}

var messageRules = [
  [/>>([0-9]+)/g, function(m) {
    var out = document.createElement('span');
    out.className = 'livechan_internallink';
    out.addEventListener('click', function() {
      var selected = document.getElementById('livechan_chat_'+m[1]);
      selected.scrollIntoView(true);
    });
    out.appendChild(document.createTextNode('>>'+m[1]));
    return out;
  }],
  [/^>.+/mg, function(m) {
    var out = document.createElement('span');
    out.className = 'livechan_greentext';
    out.appendChild(document.createTextNode(m));
    return out;
  }],
  [/\[code\]\n?([\s\S]+)\[\/code\]/g, function(m) {
    var out;
    if (m.length >= 2 && m[1].trim !== '') {
      out = document.createElement('pre');
      out.textContent = m[1];
    } else {
      out = document.createTextNode(m);
    }
    return out;
  }],
  [/\[b\]\n?([\s\S]+)\[\/b\]/g, function(m) {
    var out;
    if (m.length >= 2 && m[1].trim !== '') {
      out = document.createElement('span');
      out.className = 'livechan_boldtext';
      out.textContent = m[1];
    } else {
      out = document.createTextNode(m);
    }
    return out;
  }],
  [/\r?\n/g, function(m) {
    return document.createElement('br');
  }],

]

/* @brief Creates a chat.
 *
 * @param domElem The element to populate with chat
 *        output div and input form.
 * @param channel The channel to bind the chat to.
 */
function Chat(domElem, channel, options) {
  this.chatElems = buildChat(domElem);
  this.connection = initWebSocket(channel);
  this.initOutput();
  this.initInput();
  if (options) {
    this.options = options;
  } else {
    this.options = {};
  }
}

/* @brief Sends the message in the form.
 *
 * @param event The event causing a message to be sent.
 */
Chat.prototype.sendInput = function(event) {
  var inputElem = this.chatElems.input;
  var connection = this.connection;
  if (inputElem.message.value[0] == '/' &&
      this.options.customCommands) {
    for (var i in this.options.customCommands) {
      var regexPair = this.options.customCommands[i];
      var match = regexPair[0].exec(inputElem.message.value.slice(1));
      if (match) {
        (regexPair[1])(match);
        inputElem.message.value = '';
      }
    }
    event.preventDefault();
    return false;
  }
  if (inputElem.submit.disabled == false) {
    connection.send({
      message: inputElem.message.value,
      name: inputElem.name.value
    });
    inputElem.message.value = '';
    inputElem.submit.disabled = true;
    var i = 4;
    inputElem.submit.setAttribute('value', i);
    var countDown = setInterval(function(){
      inputElem.submit.setAttribute('value', --i);
    }, 1000);
    setTimeout(function(){
      clearInterval(countDown);
      inputElem.submit.disabled = false;
      inputElem.submit.setAttribute('value', 'send');
    }, i * 1000);
    event.preventDefault();
    return false;
  }
}

/* @brief Binds the form submission to websockets.
 */
Chat.prototype.initInput = function() {
  var inputElem = this.chatElems.input;
  var connection = this.connection;
  var self = this;
  inputElem.form.addEventListener('submit', function(event) {
    self.sendInput(event);
  });
  
  inputElem.message.addEventListener('keydown', function(event) {
    /* If enter key. */
    if (event.keyCode === 13 && !event.shiftKey) {
      self.sendInput(event);
    }
  });
}

/* @brief Binds messages to be displayed to the output.
 */
Chat.prototype.initOutput = function() {
  var outputElem = this.chatElems.output;
  var connection = this.connection;
  var self = this;
  connection.onmessage(function(data) {
    if( Object.prototype.toString.call(data) === '[object Array]' ) {
      for (var i = 0; i < data.length; i++) {
        self.insertChat(self.generateChat(data[i]), data[i].Count);
      }
    } else {
      self.insertChat(self.generateChat(data), data.Count);
    }
  });
  connection.onclose(function() {
    connection.ws = null;
    var getConnection = setInterval(function() {
      console.log("Attempting to reconnect.");
      if (initWebSocket(connection.channel, connection) !== null
          && connection.ws !== null) {
        console.log("Success!");
        clearInterval(getConnection);
      }
    }, 4000);
  });
}

/* @brief Inserts the chat into the DOM, overwriting if need be.
 *
 * @TODO: Actually scan and insert appropriately for varying numbers.
 *
 * @param outputElem The dom element to insert the chat into.
 * @param chat The dom element to be inserted.
 * @param number The number of the chat to keep it in order.
 */
Chat.prototype.insertChat = function(chat, number) {
  if (!number) {
    console.log("Error: invalid chat number.");
  }
  var outputElem = this.chatElems.output;
  var doScroll = Math.abs(outputElem.scrollTop
                 + outputElem.clientHeight
                 - outputElem.scrollHeight);
  outputElem.appendChild(chat);
  if (doScroll < 5) {
    outputElem.scrollTop = outputElem.scrollHeight;
  }
}

/* @brief Generates a chat div.
 *
 * @param data Data passed in via websocket.
 * @return A dom element.
 */
Chat.prototype.generateChat = function(data) {
  var chat = document.createElement('div');
  chat.className = 'livechan_chat_output_chat';

  var header = document.createElement('div');
  header.className = 'livechan_chat_output_header';
  var name = document.createElement('span');
  name.className = 'livechan_chat_output_name';
  var trip = document.createElement('span');
  trip.className = 'livechan_chat_output_trip';
  var date = document.createElement('span');
  date.className = 'livechan_chat_output_date';
  var count = document.createElement('span');
  count.className = 'livechan_chat_output_count';

  var body = document.createElement('div');
  body.className = 'livechan_chat_output_body';
  var message = document.createElement('div');
  message.className = 'livechan_chat_output_message';

  if (data.Name) {
    name.appendChild(document.createTextNode(data.Name));
  } else {
    name.appendChild(document.createTextNode('Anonymous'));
  }

  /* Note that parse does everything here.  If you want to change
   * how things are rendered modify messageRules. */
  if (data.Message) {
    message.appendChild(parse(data.Message, messageRules));
  } else {
    message.appendChild(document.createTextNode(''));
  }

  if (data.Date) {
    date.appendChild(document.createTextNode((new Date(data.Date)).toLocaleString()));
  }

  if (data.Trip) {
    trip.appendChild(document.createTextNode(data.Trip));
  }

  if (data.Count) {
    var self = this;
    count.setAttribute('id', 'livechan_chat_'+data.Count);
    count.appendChild(document.createTextNode(data.Count));
    count.addEventListener('click', function() {
      self.chatElems.input.message.value += '>>'+data.Count+'\n';
    });
  }

  header.appendChild(name);
  header.appendChild(trip);
  header.appendChild(date);
  header.appendChild(count);
  body.appendChild(message);

  chat.appendChild(header);
  chat.appendChild(body);
  return chat;
}

