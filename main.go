<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CHIP-8 Workstation</title>
    <style>
        :root { 
            --bg: #050505; 
            --header-bg: #ffffff;
            --accent: #22c55e;
            --border: #e5e7eb;
            --text-dim: #6b7280;
        }
        
        body {
            background: var(--bg);
            color: #111827;
            font-family: ui-monospace, 'Cascadia Code', Menlo, monospace;
            display: flex;
            flex-direction: column;
            height: 100vh;
            margin: 0;
            overflow: hidden;
        }

        /* Fixed Top Bar: Console Name + Dynamic ROM */
        .status-header {
            flex: 0 0 auto;
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 24px;
            background: var(--header-bg);
            border-bottom: 1px solid var(--border);
            font-size: 12px;
            font-weight: 700;
        }

        .brand-section { display: flex; gap: 12px; align-items: center; }
        .brand-name { font-weight: 900; color: #000; letter-spacing: -0.02em; }
        .sys-ver { color: var(--text-dim); font-weight: 400; font-size: 10px; }

        .rom-section {
            background: #f9fafb;
            padding: 4px 12px;
            border-radius: 4px;
            border: 1px solid var(--border);
        }
        
        #rom-active { color: var(--accent); text-transform: uppercase; }
        .label { color: var(--text-dim); font-weight: 400; margin-right: 4px; }

        /* Viewport: Reclaims the space for the Game */
        .viewport {
            flex: 1;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #000;
        }

        .screen-frame {
            width: 95vw;
            max-width: 1200px;
            aspect-ratio: 2 / 1;
            background: #000;
            box-shadow: 0 0 50px rgba(0,0,0,0.8);
        }

        canvas {
            width: 100% !important;
            height: 100% !important;
            image-rendering: pixelated;
        }

        /* Horizontal Legend Footer */
        .compact-footer {
            flex: 0 0 auto;
            background: #111;
            padding: 12px 24px;
            display: flex;
            justify-content: center;
            gap: 40px;
            border-top: 1px solid #222;
        }

        .legend-item { font-size: 10px; color: #888; display: flex; gap: 8px; }
        .key-tag { color: #fff; font-weight: 800; border-bottom: 1px solid #444; }
    </style>
</head>
<body>

    <header class="status-header">
        <div class="brand-section">
            <span class="brand-name">VINCENTIUS CHIP-8</span>
            <span class="sys-ver">v1.24_STABLE</span>
        </div>

        <div class="rom-section">
            <span class="label">ROM:</span>
            <span id="rom-active">INITIALIZING...</span>
        </div>

        <div class="engine-info">
            <span class="label">CORE:</span> 64X32_XOR
        </div>
    </header>

    <main class="viewport">
        <div class="screen-frame">
            <canvas id="canvas"></canvas>
        </div>
    </main>

    <footer class="compact-footer">
        <div class="legend-item">
            <span>MAP:</span>
            <span><span class="key-tag">2</span> UP</span>
            <span><span class="key-tag">Q</span> LEFT</span>
            <span><span class="key-tag">E</span> RIGHT</span>
            <span><span class="key-tag">S</span> DOWN</span>
        </div>
        <div class="legend-item">
            <span>ACTION:</span>
            <span><span class="key-tag">W</span> FIRE</span>
            <span><span class="key-tag">1-4</span> / <span class="key-tag">A-V</span> HEX_BUS</span>
        </div>
    </footer>

    <script src="wasm_glue_v1.js"></script>
    <script>
        // Update ROM name from Go
        window.updateRomName = (name) => {
            const el = document.getElementById("rom-active");
            if (el) el.innerText = name.replace('.ch8', '').toUpperCase();
        };

        const go = new Go();
        WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject).then((result) => {
            go.run(result.instance);
        });
    </script>
</body>
</html>