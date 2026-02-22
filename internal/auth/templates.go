package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Shopline CLI Setup</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-deep: #00142d;
            --bg-card: #001a3a;
            --bg-input: #002147;
            --bg-hover: #002855;
            --border: #003366;
            --border-focus: #00D4AA;
            --text: #f0f4f8;
            --text-muted: #8892a0;
            --text-dim: #4a5568;
            --accent: #00D4AA;
            --accent-secondary: #00B894;
            --accent-glow: rgba(0, 212, 170, 0.15);
            --accent-hover: #00E6B8;
            --success: #00D4AA;
            --success-glow: rgba(0, 212, 170, 0.2);
            --error: #ff6b6b;
            --error-glow: rgba(255, 107, 107, 0.15);
            --warning: #ffd93d;
            --shopline-navy: #00142d;
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
            background: var(--bg-deep);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem;
            position: relative;
            overflow-x: hidden;
        }

        /* Subtle grid pattern */
        body::before {
            content: '';
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-image:
                linear-gradient(rgba(0, 212, 170, 0.03) 1px, transparent 1px),
                linear-gradient(90deg, rgba(0, 212, 170, 0.03) 1px, transparent 1px);
            background-size: 48px 48px;
            pointer-events: none;
            z-index: 0;
        }

        /* Gradient orb effects */
        body::after {
            content: '';
            position: fixed;
            top: -30%;
            right: -20%;
            width: 80%;
            height: 80%;
            background: radial-gradient(ellipse at center, rgba(0, 212, 170, 0.08) 0%, transparent 60%);
            pointer-events: none;
            z-index: 0;
            animation: orbDrift 25s ease-in-out infinite;
        }

        @keyframes orbDrift {
            0%, 100% { transform: translate(0, 0) rotate(0deg); }
            33% { transform: translate(-5%, 8%) rotate(5deg); }
            66% { transform: translate(3%, -5%) rotate(-3deg); }
        }

        .container {
            width: 100%;
            max-width: 540px;
            position: relative;
            z-index: 1;
        }

        /* Terminal header */
        .terminal-header {
            display: flex;
            align-items: center;
            gap: 0.75rem;
            margin-bottom: 2rem;
            padding-bottom: 1.5rem;
            border-bottom: 1px solid var(--border);
        }

        .terminal-prompt {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            color: var(--text-muted);
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .terminal-prompt::before {
            content: '$';
            color: var(--accent);
            font-weight: 600;
        }

        /* Logo section */
        .logo-section {
            text-align: center;
            margin-bottom: 2.5rem;
        }

        .logo {
            height: 40px;
            margin-bottom: 1.5rem;
            display: inline-block;
            filter: brightness(0) invert(1);
            opacity: 0.95;
        }

        h1 {
            font-size: 1.875rem;
            font-weight: 700;
            letter-spacing: -0.03em;
            margin-bottom: 0.5rem;
            background: linear-gradient(135deg, var(--text) 0%, var(--text-muted) 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }

        .subtitle {
            color: var(--text-muted);
            font-size: 0.9375rem;
            font-weight: 400;
        }

        /* Card */
        .card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 20px;
            padding: 2.25rem;
            box-shadow:
                0 4px 24px rgba(0, 0, 0, 0.3),
                0 0 0 1px rgba(0, 212, 170, 0.05);
            backdrop-filter: blur(10px);
        }

        /* Form */
        .form-group {
            margin-bottom: 1.5rem;
        }

        label {
            display: block;
            font-size: 0.75rem;
            font-weight: 600;
            color: var(--text-muted);
            margin-bottom: 0.5rem;
            text-transform: uppercase;
            letter-spacing: 0.08em;
        }

        .input-wrapper {
            position: relative;
        }

        input {
            width: 100%;
            padding: 0.9375rem 1rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.9375rem;
            background: var(--bg-input);
            border: 1px solid var(--border);
            border-radius: 12px;
            color: var(--text);
            transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
        }

        input[type="password"] {
            padding-right: 3rem;
        }

        input::placeholder {
            color: var(--text-dim);
        }

        input:focus {
            outline: none;
            border-color: var(--border-focus);
            box-shadow:
                0 0 0 3px var(--accent-glow),
                inset 0 0 0 1px var(--accent);
            background: var(--bg-hover);
        }

        input:hover:not(:focus) {
            border-color: #004080;
            background: var(--bg-hover);
        }

        .toggle-visibility {
            position: absolute;
            right: 0.75rem;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            color: var(--text-dim);
            cursor: pointer;
            padding: 0.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            transition: color 0.2s;
        }

        .toggle-visibility:hover {
            color: var(--text-muted);
        }

        .input-hint {
            font-size: 0.6875rem;
            color: var(--text-dim);
            margin-top: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            line-height: 1.5;
        }

        .input-hint a {
            color: var(--accent);
            text-decoration: none;
            border-bottom: 1px dashed rgba(0, 212, 170, 0.4);
            transition: all 0.2s ease;
        }

        .input-hint a:hover {
            color: var(--accent-hover);
            border-bottom-color: var(--accent-hover);
        }

        /* API URL preview */
        .api-preview {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            color: var(--text-dim);
            background: var(--bg-input);
            padding: 0.625rem 1rem;
            border-radius: 8px;
            margin-top: 0.75rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            border: 1px solid var(--border);
        }

        .api-preview .api-icon {
            color: var(--accent);
            flex-shrink: 0;
        }

        .api-preview .api-url {
            color: var(--text-muted);
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        /* Buttons */
        .btn-group {
            display: flex;
            gap: 0.875rem;
            margin-top: 2rem;
        }

        button {
            flex: 1;
            padding: 1rem 1.5rem;
            font-family: 'Plus Jakarta Sans', sans-serif;
            font-size: 0.9375rem;
            font-weight: 600;
            border-radius: 12px;
            cursor: pointer;
            transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
            border: none;
            position: relative;
            overflow: hidden;
        }

        .btn-secondary {
            background: transparent;
            border: 1px solid var(--border);
            color: var(--text-muted);
        }

        .btn-secondary:hover:not(:disabled) {
            background: var(--bg-input);
            border-color: #004080;
            color: var(--text);
        }

        .btn-primary {
            background: linear-gradient(135deg, var(--accent) 0%, var(--accent-secondary) 100%);
            color: var(--shopline-navy);
            box-shadow:
                0 4px 20px rgba(0, 212, 170, 0.3),
                inset 0 1px 0 rgba(255, 255, 255, 0.2);
        }

        .btn-primary:hover:not(:disabled) {
            transform: translateY(-2px);
            box-shadow:
                0 8px 30px rgba(0, 212, 170, 0.4),
                inset 0 1px 0 rgba(255, 255, 255, 0.25);
        }

        .btn-primary:active:not(:disabled) {
            transform: translateY(0);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
            transform: none !important;
        }

        /* Status messages */
        .status {
            margin-top: 1.5rem;
            padding: 1rem 1.25rem;
            border-radius: 12px;
            font-size: 0.875rem;
            display: none;
            align-items: center;
            gap: 0.75rem;
            font-family: 'JetBrains Mono', monospace;
        }

        .status.show {
            display: flex;
        }

        .status.loading {
            background: var(--accent-glow);
            border: 1px solid rgba(0, 212, 170, 0.25);
            color: var(--accent);
        }

        .status.success {
            background: var(--success-glow);
            border: 1px solid rgba(0, 212, 170, 0.25);
            color: var(--success);
        }

        .status.error {
            background: var(--error-glow);
            border: 1px solid rgba(255, 107, 107, 0.25);
            color: var(--error);
        }

        .spinner {
            width: 16px;
            height: 16px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.8s linear infinite;
            flex-shrink: 0;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        /* Order count badge */
        .order-count-badge {
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 1.25rem;
            margin-top: 1.5rem;
            border-radius: 16px;
            background: linear-gradient(135deg, rgba(0, 212, 170, 0.12) 0%, rgba(0, 184, 148, 0.08) 100%);
            border: 1px solid rgba(0, 212, 170, 0.2);
            animation: slideUp 0.4s ease-out;
        }

        @keyframes slideUp {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .order-count-badge .count {
            font-size: 2.5rem;
            font-weight: 700;
            color: var(--accent);
            line-height: 1;
            margin-bottom: 0.375rem;
            font-family: 'JetBrains Mono', monospace;
        }

        .order-count-badge .label {
            font-size: 0.8125rem;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .order-count-badge .store-name {
            font-size: 0.75rem;
            color: var(--text-dim);
            margin-top: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
        }

        /* Help section */
        .help-section {
            margin-top: 2rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--border);
        }

        .help-title {
            font-size: 0.6875rem;
            font-weight: 600;
            color: var(--text-dim);
            text-transform: uppercase;
            letter-spacing: 0.1em;
            margin-bottom: 1rem;
        }

        .help-steps {
            display: flex;
            flex-direction: column;
            gap: 0.75rem;
        }

        .help-step {
            display: flex;
            align-items: flex-start;
            gap: 0.875rem;
            font-size: 0.8125rem;
            color: var(--text-muted);
        }

        .step-number {
            flex-shrink: 0;
            width: 22px;
            height: 22px;
            background: var(--bg-input);
            border: 1px solid var(--border);
            border-radius: 6px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.6875rem;
            font-weight: 600;
            color: var(--accent);
        }

        .help-step code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--bg-input);
            padding: 0.125rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            color: var(--accent);
        }

        /* Footer */
        .footer {
            text-align: center;
            margin-top: 2rem;
            font-size: 0.8125rem;
            color: var(--text-dim);
        }

        .footer a {
            color: var(--text-muted);
            text-decoration: none;
            transition: color 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }

        .footer a:hover {
            color: var(--accent);
        }

        .footer svg {
            opacity: 0.6;
            transition: opacity 0.2s;
        }

        .footer a:hover svg {
            opacity: 1;
        }

        /* Animations */
        .fade-in {
            animation: fadeIn 0.6s cubic-bezier(0.4, 0, 0.2, 1) forwards;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(12px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .card { animation-delay: 0.1s; opacity: 0; }
        .footer { animation-delay: 0.2s; opacity: 0; }

        /* Responsive */
        @media (max-width: 480px) {
            .card {
                padding: 1.5rem;
            }
            .order-count-badge .count {
                font-size: 2rem;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="terminal-header fade-in">
            <div class="terminal-prompt">
                shopline auth login
            </div>
        </div>

        <div class="logo-section fade-in">
            <svg class="logo" viewBox="0 0 151 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M0 20.3083L2.65841 16.4475C3.58135 17.4419 4.69881 18.2361 5.94146 18.7806C7.18411 19.3252 8.52543 19.6086 9.88216 19.6131C12.5781 19.6131 13.8369 18.392 13.8369 17.0393C13.8369 13 0.751497 15.8181 0.751497 7.07258C0.751497 3.21178 4.02049 0.00853409 9.37489 0.00853409C12.638 -0.112165 15.8183 1.05137 18.2331 3.24935L15.5465 6.94107C13.7549 5.2924 11.4057 4.38291 8.97097 4.39538C6.92315 4.39538 5.76772 5.33475 5.76772 6.69683C5.76772 11.1119 18.8155 8.32194 18.8155 16.5508C18.8155 20.7967 15.819 24 9.71306 24C5.27925 24 2.14176 22.5064 0 20.3083Z" fill="currentColor"/>
                <path d="M37.2554 23.5867V13.7985H26.5184V23.5867H21.6807V0.356104H26.5184V9.47737H37.2554V0.356104H42.1307V23.5867H37.2554Z" fill="currentColor"/>
                <path d="M71.5141 23.5961V0.365507H82.1196C87.0606 0.365507 89.7472 3.77541 89.7472 7.88045C89.7472 11.9855 87.0606 15.2921 82.1196 15.2921H76.3237V23.5867L71.5141 23.5961ZM84.8437 7.85227C84.8507 7.41807 84.7663 6.98724 84.5961 6.58772C84.4259 6.1882 84.1737 5.8289 83.8557 5.5331C83.5378 5.2373 83.1612 5.0116 82.7505 4.87063C82.3398 4.72966 81.904 4.67656 81.4714 4.71478H76.3612V10.9522H81.4808C81.9098 10.9902 82.3419 10.9381 82.7495 10.7992C83.1572 10.6604 83.5313 10.438 83.8479 10.1461C84.1646 9.85422 84.4167 9.49941 84.5882 9.10439C84.7597 8.70938 84.8467 8.28289 84.8437 7.85227Z" fill="currentColor"/>
                <path d="M92.5842 23.5961V0.365507H94.5568V21.7831H105.604V23.5961H92.5842Z" fill="currentColor"/>
                <path d="M107.849 23.5961V0.365507H109.793V23.5961H107.849Z" fill="currentColor"/>
                <path d="M130.027 23.6055L115.693 3.6439V23.5961H113.748V0.365507H115.721L129.971 20.0922V0.374903H131.906V23.6055H130.027Z" fill="currentColor"/>
                <path d="M135.87 23.6055V0.37491H150.29V2.18789H137.824V10.7831H150.036V12.5867H137.824V21.7831H150.299V23.5961L135.87 23.6055Z" fill="currentColor"/>
                <path d="M55.62 8.20924C55.3577 8.20738 55.1007 8.28348 54.8816 8.42788C54.6626 8.57228 54.4914 8.77849 54.3897 9.02034C54.288 9.26219 54.2604 9.5288 54.3104 9.78634C54.3605 10.0439 54.4859 10.2808 54.6707 10.467C54.8556 10.6531 55.0916 10.7802 55.3488 10.8321C55.6059 10.884 55.8727 10.8583 56.1153 10.7583C56.3579 10.6583 56.5653 10.4886 56.7113 10.2706C56.8572 10.0526 56.9351 9.79612 56.9351 9.53376C56.9352 9.1841 56.7969 8.84861 56.5505 8.60049C56.3042 8.35236 55.9697 8.21172 55.62 8.20924Z" fill="currentColor"/>
                <path d="M56.8224 0.38426C54.5535 0.38426 52.3356 1.05719 50.4492 2.31791C48.5628 3.57863 47.0927 5.3705 46.2249 7.46686C45.3571 9.56322 45.1305 11.8699 45.5738 14.095C46.0171 16.3202 47.1104 18.3639 48.7155 19.9676C50.3205 21.5713 52.365 22.6629 54.5905 23.1044C56.8161 23.5459 59.1225 23.3174 61.2182 22.4479C63.3138 21.5783 65.1045 20.1067 66.3637 18.2193C67.6229 16.3319 68.294 14.1134 68.2921 11.8446C68.2921 10.3388 67.9954 8.84776 67.4188 7.45672C66.8423 6.06568 65.9973 4.80189 64.9321 3.73758C63.867 2.67327 62.6025 1.82931 61.211 1.25393C59.8194 0.67855 58.3282 0.383025 56.8224 0.38426ZM54.5867 21.8676L52.9146 15.4423L58.2878 22.0179C57.8023 22.0857 57.3127 22.1202 56.8224 22.1212C56.0701 22.1189 55.3204 22.0339 54.5867 21.8676ZM52.708 14.4936L62.6653 20.3083C61.5037 21.1076 60.1892 21.6578 58.8045 21.924L52.708 14.4936ZM63.041 20.0358L52.5577 13.9206L50.9232 7.65497C50.8771 7.46948 50.9054 7.27333 51.0019 7.10837C51.0984 6.94341 51.2556 6.82269 51.4398 6.77197L57.6397 5.03414L66.8924 9.89067C67.025 10.5493 67.091 11.2196 67.0897 11.8915C67.0795 13.466 66.7044 15.0166 65.9937 16.4216C65.2831 17.8266 64.2563 19.0477 62.9941 19.9889L63.041 20.0358Z" fill="currentColor"/>
            </svg>
            <h1>Connect to Shopline</h1>
            <p class="subtitle">Enter your Open API credentials to get started</p>
        </div>

        <div class="card fade-in">
            <form id="setupForm" autocomplete="off">
                <div class="form-group">
                    <label for="handle">Store URL</label>
                    <input
                        type="text"
                        id="handle"
                        name="handle"
                        placeholder="https://admin.shoplineapp.com/admin/your-store/"
                        required
                        autocomplete="off"
                    >
                    <div class="input-hint">Your Shopline admin URL (copy from browser address bar)</div>
                </div>

                <div class="form-group">
                    <label for="accessToken">API Access Token</label>
                    <div class="input-wrapper">
                        <input
                            type="password"
                            id="accessToken"
                            name="accessToken"
                            placeholder="Enter your bearer token"
                            required
                            autocomplete="off"
                        >
                        <button type="button" class="toggle-visibility" onclick="togglePassword()">
                            <svg id="eyeIcon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
                                <circle cx="12" cy="12" r="3"></circle>
                            </svg>
                        </button>
                    </div>
                    <div class="input-hint">Your bearer token from the Shopline developer portal</div>
                </div>

                <div class="api-preview">
                    <svg class="api-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
                        <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
                    </svg>
                    <span class="api-url">open.shopline.io/v1</span>
                </div>

                <div class="btn-group">
                    <button type="button" id="testBtn" class="btn-secondary">Test Connection</button>
                    <button type="submit" id="submitBtn" class="btn-primary">Save &amp; Connect</button>
                </div>

                <div id="status" class="status"></div>
                <div id="orderCountBadge" class="order-count-badge" style="display: none;">
                    <span class="count" id="orderCount">0</span>
                    <span class="label">Total Orders</span>
                    <span class="store-name" id="storeBadge"></span>
                </div>
            </form>

            <div class="help-section">
                <div class="help-title">Where to find your API token</div>
                <div class="help-steps">
                    <div class="help-step">
                        <span class="step-number">1</span>
                        <span>Log in to your Shopline admin dashboard</span>
                    </div>
                    <div class="help-step">
                        <span class="step-number">2</span>
                        <span>Navigate to <code>Settings</code> &rarr; <code>Staff Settings</code></span>
                    </div>
                    <div class="help-step">
                        <span class="step-number">3</span>
                        <span>Click <code>API Auth</code> to generate your access token</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="footer fade-in">
            <a href="https://github.com/salmonumbrella/shopline-cli" target="_blank">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>

    <script>
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');
        const orderCountBadge = document.getElementById('orderCountBadge');
        const csrfToken = '{{.CSRFToken}}';

        function togglePassword() {
            const input = document.getElementById('accessToken');
            const icon = document.getElementById('eyeIcon');
            if (input.type === 'password') {
                input.type = 'text';
                icon.innerHTML = '<path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path><line x1="1" y1="1" x2="23" y2="23"></line>';
            } else {
                input.type = 'password';
                icon.innerHTML = '<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle>';
            }
        }

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            orderCountBadge.style.display = 'none';
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                const icon = type === 'success' ? '&#10003;' : '&#10007;';
                status.innerHTML = '<span style="font-size: 1.1em;">' + icon + '</span><span>' + message + '</span>';
            }
        }

        function showOrderCount(count, storeName) {
            status.className = 'status';
            orderCountBadge.style.display = 'flex';
            document.getElementById('orderCount').textContent = count.toLocaleString();
            document.getElementById('storeBadge').textContent = storeName ? 'Connected to ' + storeName : '';
        }

        function hideStatus() {
            status.className = 'status';
            orderCountBadge.style.display = 'none';
        }

        function extractStoreHandle(input) {
            // If it looks like a URL, extract the store handle from it
            const urlMatch = input.match(/admin\.shoplineapp\.com\/admin\/([^\/]+)/);
            if (urlMatch) {
                return urlMatch[1];
            }
            // Otherwise treat it as the handle directly
            return input;
        }

        function getFormData() {
            const rawHandle = document.getElementById('handle').value.trim();
            return {
                handle: extractStoreHandle(rawHandle),
                access_token: document.getElementById('accessToken').value.trim()
            };
        }

        testBtn.addEventListener('click', async () => {
            const data = getFormData();

            if (!data.handle || !data.access_token) {
                showStatus('error', 'Store URL and access token are required');
                return;
            }

            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection to Shopline API...');

            try {
                const response = await fetch('/validate', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify(data)
                });

                const result = await response.json();

                if (result.success) {
                    showOrderCount(result.order_count || 0, result.store_name);
                } else {
                    showStatus('error', result.error);
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
            }
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const data = getFormData();

            if (!data.handle || !data.access_token) {
                showStatus('error', 'Store URL and access token are required');
                return;
            }

            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');

            try {
                const response = await fetch('/submit', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'X-CSRF-Token': csrfToken
                    },
                    body: JSON.stringify(data)
                });

                const result = await response.json();

                if (result.success) {
                    showStatus('success', 'Credentials saved! Redirecting...');
                    setTimeout(() => {
                        window.location.href = '/success?store=' + encodeURIComponent(result.store_name);
                    }, 1000);
                } else {
                    showStatus('error', result.error);
                    testBtn.disabled = false;
                    submitBtn.disabled = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed: ' + err.message);
                testBtn.disabled = false;
                submitBtn.disabled = false;
            }
        });
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connected - Shopline CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-deep: #00142d;
            --bg-card: #001a3a;
            --bg-input: #002147;
            --border: #003366;
            --text: #f0f4f8;
            --text-muted: #8892a0;
            --text-dim: #4a5568;
            --accent: #00D4AA;
            --accent-glow: rgba(0, 212, 170, 0.2);
            --success: #00D4AA;
            --shopline-navy: #00142d;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: 'Plus Jakarta Sans', sans-serif;
            background: var(--bg-deep);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 2rem;
            position: relative;
        }

        body::before {
            content: '';
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background-image:
                linear-gradient(rgba(0, 212, 170, 0.02) 1px, transparent 1px),
                linear-gradient(90deg, rgba(0, 212, 170, 0.02) 1px, transparent 1px);
            background-size: 48px 48px;
            pointer-events: none;
        }

        body::after {
            content: '';
            position: fixed;
            top: 50%;
            left: 50%;
            transform: translate(-50%, -50%);
            width: 150%;
            height: 150%;
            background: radial-gradient(ellipse at center, var(--accent-glow) 0%, transparent 50%);
            pointer-events: none;
            animation: successPulse 3s ease-in-out infinite;
        }

        @keyframes successPulse {
            0%, 100% { opacity: 0.5; transform: translate(-50%, -50%) scale(1); }
            50% { opacity: 0.8; transform: translate(-50%, -50%) scale(1.1); }
        }

        .container {
            width: 100%;
            max-width: 580px;
            position: relative;
            z-index: 1;
            text-align: center;
        }

        .success-icon {
            width: 88px;
            height: 88px;
            margin: 0 auto 2rem;
            background: linear-gradient(135deg, var(--accent) 0%, #00B894 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
            box-shadow: 0 8px 40px rgba(0, 212, 170, 0.4);
        }

        .success-icon svg {
            width: 44px;
            height: 44px;
            color: var(--shopline-navy);
        }

        @keyframes scaleIn {
            from { transform: scale(0); }
            to { transform: scale(1); }
        }

        h1 {
            font-size: 2.25rem;
            font-weight: 700;
            letter-spacing: -0.03em;
            margin-bottom: 0.5rem;
            animation: fadeUp 0.5s ease 0.2s both;
        }

        .subtitle {
            color: var(--text-muted);
            font-size: 1.0625rem;
            margin-bottom: 2.5rem;
            animation: fadeUp 0.5s ease 0.3s both;
        }

        .store-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.625rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 100px;
            padding: 0.625rem 1.25rem;
            font-size: 0.9375rem;
            color: var(--text-muted);
            margin-bottom: 2.5rem;
            animation: fadeUp 0.5s ease 0.35s both;
        }

        .store-badge .dot {
            width: 10px;
            height: 10px;
            background: var(--success);
            border-radius: 50%;
            animation: dotPulse 2s ease-in-out infinite;
            box-shadow: 0 0 12px var(--accent);
        }

        @keyframes dotPulse {
            0%, 100% { opacity: 1; transform: scale(1); }
            50% { opacity: 0.7; transform: scale(0.9); }
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(12px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Terminal card */
        .terminal {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            overflow: hidden;
            text-align: left;
            animation: fadeUp 0.5s ease 0.4s both;
            box-shadow: 0 4px 32px rgba(0, 0, 0, 0.3);
        }

        .terminal-bar {
            background: var(--bg-input);
            padding: 0.875rem 1rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            border-bottom: 1px solid var(--border);
        }

        .terminal-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
        }

        .terminal-dot.red { background: #ff5f57; }
        .terminal-dot.yellow { background: #febc2e; }
        .terminal-dot.green { background: #28c840; }

        .terminal-title {
            flex: 1;
            text-align: center;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            color: var(--text-dim);
        }

        .terminal-body {
            padding: 1.5rem;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.875rem;
            margin-bottom: 1rem;
        }

        .terminal-line:last-child {
            margin-bottom: 0;
        }

        .terminal-prompt {
            color: var(--accent);
            user-select: none;
            font-weight: 600;
        }

        .terminal-text {
            color: var(--text);
        }

        .terminal-cursor {
            display: inline-block;
            width: 10px;
            height: 20px;
            background: var(--accent);
            animation: cursorBlink 1.2s step-end infinite;
            margin-left: 2px;
            vertical-align: middle;
            border-radius: 2px;
        }

        @keyframes cursorBlink {
            0%, 50% { opacity: 1; }
            50.01%, 100% { opacity: 0; }
        }

        .terminal-output {
            color: var(--success);
            padding-left: 1.25rem;
            margin-top: -0.5rem;
            margin-bottom: 1rem;
            font-size: 0.8125rem;
        }

        /* Message */
        .message {
            margin-top: 2rem;
            padding: 1.5rem;
            background: rgba(0, 212, 170, 0.08);
            border: 1px solid rgba(0, 212, 170, 0.15);
            border-radius: 16px;
            animation: fadeUp 0.5s ease 0.5s both;
        }

        .message-icon {
            font-size: 1.75rem;
            margin-bottom: 0.625rem;
        }

        .message-title {
            font-weight: 600;
            margin-bottom: 0.375rem;
            color: var(--text);
            font-size: 1.0625rem;
        }

        .message-text {
            font-size: 0.9375rem;
            color: var(--text-muted);
            line-height: 1.6;
        }

        .message-text code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--bg-input);
            padding: 0.125rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8125rem;
            color: var(--accent);
        }

        .footer {
            text-align: center;
            margin-top: 2rem;
            font-size: 0.8125rem;
            color: var(--text-dim);
            animation: fadeUp 0.5s ease 0.6s both;
        }

        .footer a {
            color: var(--text-muted);
            text-decoration: none;
            transition: color 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }

        .footer a:hover {
            color: var(--accent);
        }

        .footer svg {
            opacity: 0.6;
            transition: opacity 0.2s;
        }

        .footer a:hover svg {
            opacity: 1;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">Shopline CLI is now connected and ready to use</p>

        {{if .StoreName}}
        <div class="store-badge">
            <span class="dot"></span>
            <span>{{.StoreName}}</span>
        </div>
        {{end}}

        <div class="terminal">
            <div class="terminal-bar">
                <span class="terminal-dot red"></span>
                <span class="terminal-dot yellow"></span>
                <span class="terminal-dot green"></span>
                <span class="terminal-title">Terminal</span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-text">shopline orders list --limit 5</span>
                </div>
                <div class="terminal-output">Fetching recent orders...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-text">shopline products search "summer"</span>
                </div>
                <div class="terminal-output">Found 12 products</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            <div class="message-icon">&larr;</div>
            <div class="message-title">Return to your terminal</div>
            <div class="message-text">You can close this window and start using the CLI. Try <code>shopline --help</code> to see all available commands.</div>
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/shopline-cli" target="_blank">
                <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
        </div>
    </div>

    <script>
        // Signal completion to server
        fetch('/complete', { method: 'POST' }).catch(() => {});
    </script>
</body>
</html>`
