<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Vincentius CHIP-8</title>
    <style>
        :root { 
            --bg: #050505; 
            --header-bg: #ffffff;
            --accent: #10b981;
            --border: #e5e7eb;
            --text-dim: #6b7280;
        }
        
        body {
            background: var(--bg);
            color: #111827;
            font-family: ui-monospace, Menlo, monospace;
            display: flex;
            flex-direction: column;
            height: 100vh;
            margin: 0;
            overflow: hidden;
        }

        /* Pro Header: Branding + Dynamic Rom Name */
        .status-header {
            flex: 0 0 auto;
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 24px;
            background: var(--header-bg);
            border-bottom: 1px solid var(--border);
            font-size: 12px;
            font-weight: 700;
        }

        .brand-section { display: flex; gap: 12px; align-items: center; }
        .brand-name { font-weight: 900; color: #000; letter-spacing: -0.02em; }
        .sys-ver { color: var(--text-dim); font-weight: 400; font-size: 10px; }

        .rom-tag {
            background: #f9fafb;
            padding: 4px 12px;
            border-radius: 4px;
            border: 1px solid var(--border);
        }
        
        #rom-active { color: var(--accent); text-transform: uppercase; }
        .label { color: var(--text-dim); font-weight: 400; margin-right: 4px; }

        /* Viewport: Reclaims the space for the Game */
        .viewport {
            flex: 1; /* Takes almost all screen space */
            display: flex;
            align-items: center;
            justify-content: center;
            background: #000;
            padding: 10px;
        }

        .screen-frame {
            width: 100%;
            max-width: 1200px;
            aspect-ratio: 2 / 1;
            background: #000;
        }

        canvas {
            width: 100% !important;
            height: 100% !important;
            image-rendering: pixelated; /* Sharp pixels for Test Suite */
        }

        /* Minimal Key Legend Footer */
        .compact-footer {
            flex: 0 0 auto;
            background: #111;
            padding: 10px 24px;
            display: flex;
            justify-content: center;
            gap: 30px;
            border-top: 1px solid #222;
        }

        .legend-group { font-size: 10px; color: #888; display: flex; gap: 6px; align-items: center; }
        .key-tag { color: #fff; font-weight: 800; padding: 1px 4px; background: #333; border-radius: 2px; }
    </style>
</head>
<body>

    <header class="status-header">
        <div class="brand-section">
            <span class="brand-name">VINCENTIUS WORKSTATION</span>
            <span class="sys-ver">CORE_V1.24</span>
        </div>

        <div class="rom-tag">
            <span class="label">ROM:</span>
            <span id="rom-active">BOOTING...</span>
        </div>

        <div class="engine-info">
            <span class="label">ENGINE:</span> 64X32_XOR
        </div>
    </header>

    <main class="viewport">
        <div class="screen-frame">
            <canvas id="canvas"></canvas>
        </div>
    </main>

    <footer class="compact-footer">
        <div class="legend-group">
            <span>MOV:</span>
            <span><span class="key-tag">2</span> UP</span>
            <span><span class="key-tag">Q</span> LEFT</span>
            <span><span class="key-tag">E</span> RIGHT</span>
            <span><span class="key-tag">S</span> DOWN</span>
        </div>
        <div class="legend-group">
            <span>ACTION:</span>
            <span><span class="key-tag">W</span> FIRE</span>
            <span><span class="key-tag">1-4 / A-V</span> HEX_KEYS</span>
        </div>
    </footer>

    <script src="wasm_glue_v1.js"></script>
    <script>
        // Update ROM name dynamically from Go
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