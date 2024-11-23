function syntaxHighlight(json) {
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:\s*)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        let cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        switch (cls) {
            case 'key':
                return `\x1b[33m${match}\x1b[0m`; // Yellow
            case 'string':
                return `\x1b[32m${match}\x1b[0m`; // Green
            case 'number':
                return `\x1b[35m${match}\x1b[0m`; // Magenta
            case 'boolean':
                return `\x1b[36m${match}\x1b[0m`; // Cyan
            case 'null':
                return `\x1b[31m${match}\x1b[0m`; // Red
            default:
                return match;
        }
    });
  }
  function getRandomColor() {
    const colors = [
        '\x1b[31m', // Red
        '\x1b[32m', // Green
        '\x1b[33m', // Yellow
        '\x1b[34m', // Blue
        '\x1b[35m', // Magenta
        '\x1b[36m', // Cyan
    ];
    return colors[Math.floor(Math.random() * colors.length)];
  }
  const stateName = {
    0: "Idle",
    1: "Initialized",
    2: "Running",
    3: "Connected",
    4: "Syncing",
  };
  
  export { syntaxHighlight , getRandomColor , stateName };