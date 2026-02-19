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
      .toolbar select,
      .toolbar input {
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
          Examples
          <select id="example-picker">
            <option value="">choose an example</option>
          </select>
        </label>
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
        <label>
          <input id="pretty-print" type="checkbox" checked />
          Pretty print
        </label>
        <label>
          API key
          <input id="api-key" type="password" placeholder="optional" />
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
      const examplePicker = document.getElementById("example-picker");
      const outputFormat = document.getElementById("output-format");
      const prettyPrint = document.getElementById("pretty-print");
      const apiKey = document.getElementById("api-key");
      const scriptArea = document.getElementById("script");

      const examples = [
        {
          id: "revenue-dashboard",
          name: "Revenue dashboard (grouping + maxBy)",
          output: "",
          script:
            "%im 0.1\n" +
            "fun sum(nums) = nums reduce (acc, n) -> acc + n\n" +
            "output application/json\n" +
            "---\n" +
            "{\n" +
            "  totals: {\n" +
            "    orders: sizeOf(payload),\n" +
            "    revenue: sum(payload map (o) -> o.total)\n" +
            "  },\n" +
            "  byRegion: payload\n" +
            "    groupBy (o) -> o.region\n" +
            "    mapObject (v, k) -> [k, {\n" +
            "      orders: sizeOf(v),\n" +
            "      revenue: sum(v map (o) -> o.total),\n" +
            "      topSku: (v maxBy (o) -> o.total).sku\n" +
            "    }],\n" +
            "  flagged: payload filter (o) -> o.total > 1000\n" +
            "}\n",
          inputs: [
            {
              name: "payload",
              format: "json",
              content:
                "[\n" +
                "  {\"id\": \"A100\", \"region\": \"North\", \"sku\": \"BK-42\", \"total\": 2200},\n" +
                "  {\"id\": \"A101\", \"region\": \"North\", \"sku\": \"BK-17\", \"total\": 180},\n" +
                "  {\"id\": \"B200\", \"region\": \"West\", \"sku\": \"GL-55\", \"total\": 1250},\n" +
                "  {\"id\": \"C300\", \"region\": \"West\", \"sku\": \"GL-55\", \"total\": 480},\n" +
                "  {\"id\": \"D400\", \"region\": \"East\", \"sku\": \"ST-01\", \"total\": 640}\n" +
                "]\n",
            },
          ],
        },
        {
          id: "inventory-csv",
          name: "Inventory export to CSV",
          output: "",
          script:
            "%im 0.1\n" +
            "output text/csv\n" +
            "---\n" +
            "payload map (p) -> {\n" +
            "  sku: p.sku,\n" +
            "  status: if (p.onHand > p.reorderPoint) \"ok\" else \"reorder\",\n" +
            "  value: p.onHand * p.unitCost\n" +
            "}\n",
          inputs: [
            {
              name: "payload",
              format: "json",
              content:
                "[\n" +
                "  {\"sku\": \"BK-42\", \"onHand\": 120, \"reorderPoint\": 50, \"unitCost\": 12.5},\n" +
                "  {\"sku\": \"GL-55\", \"onHand\": 20, \"reorderPoint\": 35, \"unitCost\": 48.0},\n" +
                "  {\"sku\": \"ST-01\", \"onHand\": 8, \"reorderPoint\": 10, \"unitCost\": 210.0}\n" +
                "]\n",
            },
          ],
        },
        {
          id: "customer-merge",
          name: "Customer 360 (merge two inputs)",
          output: "",
          script:
            "%im 0.1\n" +
            "output application/json\n" +
            "---\n" +
            "profiles map (p) ->\n" +
            "  p ++ {\n" +
            "    lastLogin: (activity filter (a) -> a.userId == p.id)[0].lastLogin,\n" +
            "    events: sizeOf(activity filter (a) -> a.userId == p.id)\n" +
            "  }\n",
          inputs: [
            {
              name: "profiles",
              format: "json",
              content:
                "[\n" +
                "  {\"id\": 101, \"name\": \"Ava\", \"tier\": \"gold\"},\n" +
                "  {\"id\": 102, \"name\": \"Noah\", \"tier\": \"silver\"},\n" +
                "  {\"id\": 103, \"name\": \"Mia\", \"tier\": \"gold\"}\n" +
                "]\n",
            },
            {
              name: "activity",
              format: "json",
              content:
                "[\n" +
                "  {\"userId\": 101, \"lastLogin\": \"2026-01-30T09:12:00Z\"},\n" +
                "  {\"userId\": 101, \"lastLogin\": \"2026-01-31T10:44:00Z\"},\n" +
                "  {\"userId\": 102, \"lastLogin\": \"2026-01-25T08:05:00Z\"},\n" +
                "  {\"userId\": 103, \"lastLogin\": \"2026-01-29T16:32:00Z\"},\n" +
                "  {\"userId\": 103, \"lastLogin\": \"2026-02-01T18:20:00Z\"}\n" +
                "]\n",
            },
          ],
        },
      ];

      function createInputCard(index, preset) {
        const card = document.createElement("div");
        card.className = "input-card";
        const defaultName =
          preset && preset.name
            ? preset.name
            : index === 0
            ? "payload"
            : "input" + (index + 1);

        const nameLabel = document.createElement("label");
        nameLabel.appendChild(document.createTextNode("Input name "));
        const nameInput = document.createElement("input");
        nameInput.type = "text";
        nameInput.value = defaultName;
        nameLabel.appendChild(nameInput);

        const formatLabel = document.createElement("label");
        formatLabel.appendChild(document.createTextNode("Format"));
        const formatSelect = document.createElement("select");
        const formatOptions = [
          { value: "", label: "auto" },
          { value: "json", label: "json" },
          { value: "xml", label: "xml" },
          { value: "csv", label: "csv" },
          { value: "yaml", label: "yaml" },
          { value: "text/plain", label: "text" },
        ];
        formatOptions.forEach((item) => {
          const option = document.createElement("option");
          option.value = item.value;
          option.textContent = item.label;
          formatSelect.appendChild(option);
        });
        formatLabel.appendChild(formatSelect);

        const contentLabel = document.createElement("label");
        contentLabel.appendChild(document.createTextNode("Content "));
        const contentInput = document.createElement("textarea");
        contentInput.placeholder = "Paste input data here";
        contentLabel.appendChild(contentInput);

        card.appendChild(nameLabel);
        card.appendChild(formatLabel);
        card.appendChild(contentLabel);
        if (preset) {
          card.querySelector("select").value = preset.format || "";
          card.querySelector("textarea").value = preset.content || "";
        }
        return card;
      }

      function addInputCard(preset) {
        const card = createInputCard(inputsList.children.length, preset);
        inputsList.appendChild(card);
      }

      function setInputs(exampleInputs) {
        inputsList.innerHTML = "";
        if (exampleInputs && exampleInputs.length) {
          exampleInputs.forEach((input, index) => {
            inputsList.appendChild(createInputCard(index, input));
          });
          return;
        }
        addInputCard();
      }

      function resetOutput() {
        status.textContent = "";
        result.value = "";
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

      function applyExample(example) {
        if (!example) {
          return;
        }
        scriptArea.value = example.script || "";
        outputFormat.value = example.output || "";
        setInputs(example.inputs);
        resetOutput();
      }

      function populateExamples() {
        examples.forEach((example) => {
          const option = document.createElement("option");
          option.value = example.id;
          option.textContent = example.name;
          examplePicker.appendChild(option);
        });
      }

      function onExampleChange() {
        const selected = examples.find((example) => example.id === examplePicker.value);
        applyExample(selected);
      }

      function normalizeMimeType(format, fallbackMimeType) {
        if (format) {
          if (format.indexOf("/") >= 0) {
            return format;
          }
          const shortToMime = {
            json: "application/json",
            xml: "application/xml",
            csv: "application/csv",
            yaml: "application/yaml",
            text: "text/plain",
          };
          return shortToMime[format] || format;
        }
        if (!fallbackMimeType) {
          return "";
        }
        return fallbackMimeType.split(";")[0].trim();
      }

      function prettyPrintXML(xml) {
        const normalized = xml.trim().replace(/>\s*</g, "><");
        const lines = normalized.replace(/(>)(<)(\/*)/g, "$1\n$2$3").split("\n");
        let indent = 0;
        const formatted = [];
        lines.forEach((line) => {
          if (!line) {
            return;
          }
          if (line.match(/^<\/.+>/)) {
            indent = Math.max(indent - 1, 0);
          }
          formatted.push("  ".repeat(indent) + line);
          if (line.match(/^<[^!?/][^>]*[^/]>/)) {
            indent++;
          }
        });
        return formatted.join("\n");
      }

      function formatResultOutput(raw, mimeType) {
        if (!prettyPrint.checked) {
          return raw;
        }
        if (mimeType === "application/json") {
          try {
            return JSON.stringify(JSON.parse(raw), null, 2);
          } catch (_) {
            return raw;
          }
        }
        if (mimeType === "application/xml" || mimeType === "text/xml") {
          return prettyPrintXML(raw);
        }
        return raw;
      }

      async function runScript() {
        const script = scriptArea.value;
        const output = outputFormat.value;
        const payload = {
          script,
          output,
          inputs: collectInputs(),
        };
        status.textContent = "Running...";
        result.value = "";
        try {
          if (window.infomungeRun) {
            const response = window.infomungeRun(JSON.stringify(payload));
            const responseText = response && response.result ? response.result : "";
            const responseMimeType = response && response.mimeType ? response.mimeType : "";
            const mimeType = normalizeMimeType(output, responseMimeType);
            result.value = formatResultOutput(responseText, mimeType);
            if (response && response.ok) {
              status.textContent = "Success";
            } else {
              const message = response && response.error ? response.error : "Unknown error";
              status.textContent = "Error";
              result.value = message;
            }
            return;
          }
          const headers = { "Content-Type": "application/json" };
          if (apiKey.value) {
            headers["X-API-Key"] = apiKey.value;
          }
          const response = await fetch("/run", {
            method: "POST",
            headers,
            body: JSON.stringify(payload),
          });
          const text = await response.text();
          const mimeType = normalizeMimeType(output, response.headers.get("Content-Type") || "");
          result.value = formatResultOutput(text, mimeType);
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
      examplePicker.addEventListener("change", onExampleChange);
      populateExamples();
      addInputCard();
    </script>
  </body>
</html>`
