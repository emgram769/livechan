/* @file chat.js
 * 
 * Include this file to spawn a livechan chat.
 * Use createChat(domElement, channel).
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
      file: file
    }
  };
}

function Connection(ws) {
  this.ws = ws;
}

Connection.prototype.send = function(obj) {
  /* Jsonify the object and send as string. */
  this.ws.send(JSON.stringify(obj));
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
function initWebSocket(channel) {
  var ws = null;
  if (window['WebSocket']) {
    ws = new WebSocket('ws://'+location.host+'/'+channel);
  }
  if (ws !== null) {
    return new Connection(ws);
  } else {
    return null;
  }
}

/* @brief Inserts the chat into the DOM, overwriting if need be.
 *
 * @param outputElem The dom element to insert the chat into.
 * @param chat The dom element to be inserted.
 * @param number The number of the chat to keep it in order.
 */
function insertChat(outputElem, chat, number) {
  var doScroll = outputElem.scrollTop + outputElem.clientHeight == outputElem.scrollHeight;
  outputElem.appendChild(chat);
  if (doScroll) {
    outputElem.scrollTop = outputElem.scrollHeight;
  }
}

/* @brief Generates a chat div.
 *
 * @param data Data passed in via websocket.
 * @return A dom element.
 */
function generateChat(data) {
  var chat = document.createElement('div');
  chat.className = 'livechan_chat_output_chat';
  var name = document.createElement('div');
  name.className = 'livechan_chat_output_name';
  var message = document.createElement('div');
  message.className = 'livechan_chat_output_message';

  if (data.name) {
    name.appendChild(document.createTextNode(data.name));
  } else {
    name.appendChild(document.createTextNode('Anonymous'));
  }

  /* TODO: actually do all the processing to make it a real message. */
  if (data.message) {
    message.appendChild(document.createTextNode(data.message));
  } else {
    message.appendChild(document.createTextNode(''));
  }

  chat.appendChild(name);
  chat.appendChild(message);
  return chat;
}

/* @brief Binds messages to be displayed to the output.
 *
 * @param outputElem The DOM element to be populated
          with messages.
 * @param connection The websocket connection.
 */
function initOutput(outputElem, connection) {
  connection.onmessage(function(data) {
    insertChat(outputElem, generateChat(data));
  });
}

/* @brief Binds the form submission to websockets.
 *
 * @param inputElem The form itself.
 * @param connection The websocket connection.
 */
function initInput(inputElem, connection) {
  inputElem.form.addEventListener('submit', function(event) {
    connection.send({
      message: inputElem.message.value,
      name: inputElem.name.value
    });
    event.preventDefault();
    inputElem.message.value = '';
    return false;
  });
  
  inputElem.message.addEventListener('keydown', function(event) {
    /* If enter key. */
    if (event.keyCode === 13 && !event.shiftKey) {
      event.preventDefault();
      connection.send({
        message: inputElem.message.value,
        name: inputElem.name.value
      });
      inputElem.message.value = '';
      return false;
    }
  });
}

/* @brief Creates a chat.
 *
 * @param domElem The element to populate with chat
 *        output div and input form.
 * @param channel The channel to bind the chat to.
 */
function createChat(domElem, channel) {
  var chatElems = buildChat(domElem);
  var connection = initWebSocket(channel);
  initInput(chatElems.input, connection);
  initOutput(chatElems.output, connection);
}

