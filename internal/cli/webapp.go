package cli

const playgroundHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>InfoMunge Playground</title>
    <style>
      :root {
        --ink: #1f1b16;
        --accent: #d56733;
        --bg: #f4efe9;
        --panel: #fffdf8;
        --border: #e3d9cd;
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        font-family: "Georgia", "Times New Roman", serif;
        color: var(--ink);
        background: radial-gradient(circle at top left, #fff8ef 0%, var(--bg) 45%, #f1e7dc 100%);
      }
      .app {
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        padding: 24px;
        gap: 16px;
      }
      header {
        display: flex;
        align-items: baseline;
        justify-content: space-between;
        gap: 12px;
      }
      header h1 {
        margin: 0;
        font-size: 28px;
        letter-spacing: 0.5px;
      }
      header p {
        margin: 0;
        font-size: 14px;
        opacity: 0.7;
      }
      .toolbar {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        align-items: center;
      }
      .toolbar button {
        background: var(--accent);
        color: white;
        border: none;
        padding: 10px 18px;
        font-size: 14px;
        border-radius: 18px;
        cursor: pointer;
      }
      .toolbar button:hover {
        opacity: 0.9;
      }
      .toolbar select {
        border: 1px solid var(--border);
        padding: 8px 10px;
        border-radius: 12px;
        background: var(--panel);
      }
      .layout {
        flex: 1;
        display: grid;
        grid-template-columns: minmax(220px, 1fr) minmax(260px, 1.2fr) minmax(220px, 1fr);
        gap: 16px;
      }
      .panel {
        background: var(--panel);
        border: 1px solid var(--border);
        border-radius: 18px;
        padding: 16px;
        display: flex;
        flex-direction: column;
        gap: 12px;
        min-height: 0;
      }
      .panel h2 {
        margin: 0;
        font-size: 16px;
        letter-spacing: 0.6px;
        text-transform: uppercase;
      }
      .inputs-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }
      .inputs-header button {
        border: 1px dashed var(--border);
        background: transparent;
        padding: 6px 12px;
        border-radius: 12px;
        cursor: pointer;
      }
      .input-card {
        border: 1px solid var(--border);
        border-radius: 14px;
        padding: 10px;
        display: grid;
        gap: 8px;
      }
      .input-card input,
      .input-card select,
      .input-card textarea,
      #script,
      #result {
        width: 100%;
        border: 1px solid var(--border);
        border-radius: 10px;
        padding: 8px;
        font-family: "Courier New", monospace;
        font-size: 13px;
        background: #fffefb;
      }
      .input-card textarea {
        min-height: 110px;
        resize: vertical;
      }
      #script {
        flex: 1;
        min-height: 260px;
        resize: vertical;
      }
      #result {
        flex: 1;
        min-height: 260px;
        resize: none;
        background: #f9f4ee;
      }
      .status {
        font-size: 13px;
        color: #7b4b32;
      }
      @media (max-width: 980px) {
        .layout {
          grid-template-columns: 1fr;
        }
      }
    </style>
  </head>
  <body>
    <div class="app">
      <header>
        <div>
          <h1>InfoMunge Playground</h1>
          <p>Draft inputs, write scripts, and see results instantly.</p>
        </div>
        <p>Server mode • /run endpoint</p>
      </header>
      <div class="toolbar">
        <button id="run-script" type="button">Run Script</button>
        <label>
          Output format
          <select id="output-format">
            <option value="">auto (script header)</option>
            <option value="json">json</option>
            <option value="xml">xml</option>
            <option value="csv">csv</option>
            <option value="yaml">yaml</option>
            <option value="text/plain">text</option>
          </select>
        </label>
      </div>
      <main class="layout">
        <section class="panel" id="inputs-panel">
          <div class="inputs-header">
            <h2>Inputs</h2>
            <button id="add-input" type="button">+ Add input</button>
          </div>
          <div id="inputs-list"></div>
        </section>
        <section class="panel" id="script-panel">
          <h2>Script</h2>
          <textarea id="script">%im 0.1
output application/json
---
payload</textarea>
        </section>
        <section class="panel" id="result-panel">
          <h2>Result</h2>
          <textarea id="result" readonly></textarea>
          <div class="status" id="status"></div>
        </section>
      </main>
    </div>
    <script>
      const inputsList = document.getElementById("inputs-list");
      const status = document.getElementById("status");
      const result = document.getElementById("result");

      function createInputCard(index) {
        const card = document.createElement("div");
        card.className = "input-card";
        const defaultName = index === 0 ? "payload" : "input" + (index + 1);
        card.innerHTML =
          '<label>Input name <input type="text" value="' + defaultName + '" /></label>' +
          '<label>Format' +
          '  <select>' +
          '    <option value="">auto</option>' +
          '    <option value="json">json</option>' +
          '    <option value="xml">xml</option>' +
          '    <option value="csv">csv</option>' +
          '    <option value="yaml">yaml</option>' +
          '    <option value="text/plain">text</option>' +
          '  </select>' +
          '</label>' +
          '<label>Content <textarea placeholder="Paste input data here"></textarea></label>';
        return card;
      }

      function addInputCard() {
        const card = createInputCard(inputsList.children.length);
        inputsList.appendChild(card);
      }

      function collectInputs() {
        const inputs = [];
        Array.from(inputsList.children).forEach((card) => {
          const name = card.querySelector("input").value.trim();
          const format = card.querySelector("select").value;
          const content = card.querySelector("textarea").value;
          if (!name) {
            return;
          }
          inputs.push({
            name,
            format: format || undefined,
            content,
          });
        });
        return inputs;
      }

      async function runScript() {
        const script = document.getElementById("script").value;
        const output = document.getElementById("output-format").value;
        status.textContent = "Running...";
        result.value = "";
        try {
          const response = await fetch("/run", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              script,
              output,
              inputs: collectInputs(),
            }),
          });
          const text = await response.text();
          result.value = text;
          status.textContent = response.ok ? "Success" : "Error";
          if (!response.ok) {
            status.textContent += " " + response.status;
          }
        } catch (err) {
          status.textContent = "Error: " + err;
        }
      }

      document.getElementById("run-script").addEventListener("click", runScript);
      document.getElementById("add-input").addEventListener("click", addInputCard);
      addInputCard();
    </script>
  </body>
</html>`
