document.addEventListener('DOMContentLoaded', () => {
    const API_BASE_URL = '/api'; // 相対パスに変更 (CORSの問題を回避)
    let serviceRunning = false;
    let lastCheckTime = null;

    // --- Element References ---
    const startServiceBtn = document.getElementById('startServiceBtn');
    const stopServiceBtn = document.getElementById('stopServiceBtn');
    const getServiceStatusBtn = document.getElementById('getServiceStatusBtn');
    const serviceStatusSpan = document.getElementById('serviceStatus');
    const statusDotIndicator = document.getElementById('status-dot-indicator');
    const statusLineIndicator = document.getElementById('status-line-indicator');

    const getConfigBtn = document.getElementById('getConfigBtn');
    const saveConfigBtn = document.getElementById('saveConfigBtn');
    const configDisplay = document.getElementById('configDisplay');
    const updateConfigBtn = document.getElementById('updateConfigBtn');

    const getDevicesBtn = document.getElementById('getDevicesBtn');
    const devicesListDiv = document.getElementById('devicesList');
    const keyboardDeviceSelect = document.getElementById('keyboardDeviceSelect');
    const mouseDeviceSelect = document.getElementById('mouseDeviceSelect');
    const setPreferredDevicesBtn = document.getElementById('setPreferredDevicesBtn');

    const healthCheckBtn = document.getElementById('healthCheckBtn');
    const healthStatusSpan = document.getElementById('healthStatus');
    const healthDotIndicator = document.getElementById('health-dot-indicator');
    const lastCheckedSpan = document.getElementById('lastChecked');

    // --- 洗練されたトースト通知システム ---
    function showToast(message, type = 'info', duration = 3000) {
        const existingToast = document.querySelector('.toast');
        if (existingToast) {
            existingToast.remove();
        }

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        // アイコンを追加
        let icon = '';
        if (type === 'success') icon = '<i class="fas fa-check-circle mr-2"></i>';
        else if (type === 'error') icon = '<i class="fas fa-exclamation-circle mr-2"></i>';
        else if (type === 'info') icon = '<i class="fas fa-info-circle mr-2"></i>';
        
        toast.innerHTML = `${icon}${message}`;
        document.body.appendChild(toast);

        // アニメーション付きで表示
        setTimeout(() => {
            toast.classList.add('show');
            setTimeout(() => {
                toast.classList.remove('show');
                setTimeout(() => toast.remove(), 300);
            }, duration);
        }, 10);
    }

    // --- 日付フォーマット関数 ---
    function formatDateTime(date) {
        const options = {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        };
        return new Date(date).toLocaleString('ja-JP', options);
    }

    // --- ヘルパー関数の強化 ---
    async function apiRequest(endpoint, method = 'GET', body = null) {
        const btnSelector = getButtonSelectorForEndpoint(endpoint, method);
        const btn = document.querySelector(btnSelector);

        // startTime変数をスコープの外で宣言
        const startTime = Date.now();

        if (btn) {
            const originalContent = btn.innerHTML;
            btn.innerHTML = `<span class="loading-spinner mr-2"></span>`;
            btn.disabled = true;
            
            // スムーズなアニメーション効果のため、少なくとも300ms待機
        }

        try {
            const options = {
                method: method,
                headers: {},
            };

            if (body) {
                options.headers['Content-Type'] = 'application/json';
                options.body = JSON.stringify(body);
            }

            const response = await fetch(`${API_BASE_URL}${endpoint}`, options);

            if (!response.ok) {
                let errorMessage;
                try {
                    const errorData = await response.json();
                    errorMessage = errorData.error || `サーバーエラー: ${response.status}`;
                } catch (e) {
                    errorMessage = `HTTP エラー: ${response.status}`;
                }
                throw new Error(errorMessage);
            }

            if (response.status === 204 || (response.headers.get('content-length') === '0')) {
                return { status: 'success' };
            }

            return await response.json();

        } catch (error) {
            console.error('API リクエストエラー:', error);
            showToast(error.message, 'error');
            throw error;
        } finally {
            if (btn) {
                // 少なくとも300ms経過するまで待機して、アニメーションが見えるようにする
                const elapsedTime = Date.now() - startTime;
                const remainingTime = Math.max(0, 300 - elapsedTime);
                
                setTimeout(() => {
                    setButtonIcon(btn, endpoint, method);
                    btn.disabled = false;
                    
                    // 成功時の軽い春のような効果 (Appleデバイスの触覚フィードバックを視覚的に再現)
                    btn.classList.add('button-spring');
                    setTimeout(() => btn.classList.remove('button-spring'), 300);
                }, remainingTime);
            }
        }
    }

    function getButtonSelectorForEndpoint(endpoint, method) {
        if (endpoint === '/service/start' && method === 'POST') return '#startServiceBtn';
        if (endpoint === '/service/stop' && method === 'POST') return '#stopServiceBtn';
        if (endpoint === '/service/status' && method === 'GET') return '#getServiceStatusBtn';
        if (endpoint === '/config' && method === 'GET') return '#getConfigBtn';
        if (endpoint === '/config' && method === 'PUT') return '#updateConfigBtn';
        if (endpoint === '/config/save' && method === 'POST') return '#saveConfigBtn';
        if (endpoint === '/devices' && method === 'GET') return '#getDevicesBtn';
        if (endpoint === '/devices/preferred' && method === 'PUT') return '#setPreferredDevicesBtn';
        if (endpoint === '/health' && method === 'GET') return '#healthCheckBtn';

        return null;
    }

    function setButtonIcon(btn, endpoint, method) {
        if (endpoint === '/service/start' && method === 'POST') {
            btn.innerHTML = `<i class="fas fa-play"></i>`;
        } else if (endpoint === '/service/stop' && method === 'POST') {
            btn.innerHTML = `<i class="fas fa-stop"></i>`;
        } else if (endpoint === '/service/status' && method === 'GET') {
            btn.innerHTML = `<i class="fas fa-sync-alt"></i>`;
        } else if (endpoint === '/config' && method === 'GET') {
            btn.innerHTML = `<i class="fas fa-download"></i>`;
        } else if (endpoint === '/config' && method === 'PUT') {
            btn.innerHTML = `<i class="fas fa-upload mr-2"></i>`;
        } else if (endpoint === '/config/save' && method === 'POST') {
            btn.innerHTML = `<i class="fas fa-save"></i>`;
        } else if (endpoint === '/devices' && method === 'GET') {
            btn.innerHTML = `<i class="fas fa-sync-alt"></i>`;
        } else if (endpoint === '/devices/preferred' && method === 'PUT') {
            btn.innerHTML = `<i class="fas fa-save mr-2"></i>`;
        } else if (endpoint === '/health' && method === 'GET') {
            btn.innerHTML = `<i class="fas fa-stethoscope"></i>`;
        } else {
            btn.innerHTML = '';
        }
    }

    // --- ステータス表示の更新 ---
    function updateStatusIndicators(status) {
        // テキスト表示
        serviceStatusSpan.textContent = status;
        
        // ステータスドット
        statusDotIndicator.className = 'status-dot';
        if (status.toLowerCase() === 'running') {
            statusDotIndicator.classList.add('running');
        } else if (status.toLowerCase() === 'stopped' || status.toLowerCase() === 'error') {
            statusDotIndicator.classList.add('error');
        } else {
            statusDotIndicator.classList.add('idle');
        }
        
        // ステータスライン
        statusLineIndicator.className = 'w-20 status-line';
        if (status.toLowerCase() === 'running') {
            statusLineIndicator.classList.add('running');
        } else if (status.toLowerCase() === 'stopped' || status.toLowerCase() === 'error') {
            statusLineIndicator.classList.add('error');
        } else {
            statusLineIndicator.classList.add('partial');
        }
    }
    
    // ヘルスステータス更新
    function updateHealthIndicators(status) {
        const statusText = healthStatusSpan.querySelector('span:not(.status-dot)');
        if (statusText) {
            statusText.textContent = status;
        }
        
        // ヘルスドット
        healthDotIndicator.className = 'status-dot';
        if (status.toLowerCase() === 'ok') {
            healthDotIndicator.classList.add('running');
        } else if (status.toLowerCase() === 'error') {
            healthDotIndicator.classList.add('error');
        } else {
            healthDotIndicator.classList.add('idle');
        }
    }

    // --- サービス管理機能の強化 ---
    startServiceBtn.addEventListener('click', async () => {
        try {
            const data = await apiRequest('/service/start', 'POST');
            if (data.status === 'already_running') {
                showToast('サービスはすでに実行中です', 'info');
            } else {
                showToast('サービスを開始しました', 'success');
            }
            getServiceStatus();
        } catch (error) {}
    });

    stopServiceBtn.addEventListener('click', async () => {
        try {
            const data = await apiRequest('/service/stop', 'POST');
            if (data.status === 'not_running') {
                showToast('サービスは既に停止しています', 'info');
            } else {
                showToast('サービスを停止しました', 'success');
            }
            getServiceStatus();
        } catch (error) {}
    });

    async function getServiceStatus() {
        try {
            const data = await apiRequest('/service/status');
            updateStatusIndicators(data.status);
            serviceRunning = data.status === 'running';

            startServiceBtn.disabled = serviceRunning;
            stopServiceBtn.disabled = !serviceRunning;
        } catch (error) {
            updateStatusIndicators('Error');
        }
    }
    getServiceStatusBtn.addEventListener('click', getServiceStatus);

    // --- Configuration Management ---
    async function getConfig() {
        try {
            const data = await apiRequest('/config');
            configDisplay.value = JSON.stringify(data, null, 2);
            showToast('設定を取得しました', 'info');
        } catch (error) {
            configDisplay.value = `Error: ${error.message}`;
        }
    }
    getConfigBtn.addEventListener('click', getConfig);

    saveConfigBtn.addEventListener('click', async () => {
        try {
            const data = await apiRequest('/config/save', 'POST', { path: "" });
            showToast(`設定を保存しました: ${data.path || 'デフォルト場所'}`, 'success');
        } catch (error) {}
    });

    updateConfigBtn.addEventListener('click', async () => {
        try {
            let configData;
            try {
                configData = JSON.parse(configDisplay.value);
            } catch (parseError) {
                showToast('エラー: JSONの形式が正しくありません', 'error');
                return;
            }

            const data = await apiRequest('/config', 'PUT', configData);
            showToast('設定を更新しました', 'success');
        } catch (error) {}
    });

    // --- Device Management ---
    async function getDevices() {
        try {
            console.log("デバイス情報を取得中...");
            const devices = await apiRequest('/devices');
            console.log("取得したデバイス:", devices); 

            // リストとセレクトボックスをリセット
            devicesListDiv.innerHTML = '';
            keyboardDeviceSelect.innerHTML = '<option value="">-- キーボードを選択 --</option>';
            mouseDeviceSelect.innerHTML = '<option value="">-- マウスを選択 --</option>';

            // 現在の設定を取得して優先デバイスを特定
            const currentConfig = await apiRequest('/config');
            console.log("現在の設定:", currentConfig);
            const preferredKeyboard = currentConfig.device_prefs?.preferred_keyboard_device || "";
            const preferredMouse = currentConfig.device_prefs?.preferred_mouse_device || "";

            if (!devices || devices.length === 0) {
                devicesListDiv.innerHTML = '<p class="text-gray-500 text-center">検出されたデバイスがありません</p>';
                return;
            }

            devices.forEach(device => {
                // デバイスタイプを数値から文字列へ正しく変換
                let deviceTypeStr;
                // Type: 0 = Keyboard, 1 = Mouse (DeviceTypeKeyboard, DeviceTypeMouse)
                if (device.Type === 0) {
                    deviceTypeStr = 'keyboard';
                } else if (device.Type === 1) {
                    deviceTypeStr = 'mouse';
                } else {
                    deviceTypeStr = 'other';
                }

                // デバイスリストにアイテムを追加
                const deviceElement = document.createElement('div');
                deviceElement.className = `device-item ${deviceTypeStr}`;

                let typeIcon = 'fas fa-question-circle';
                if (deviceTypeStr === 'keyboard') typeIcon = 'fas fa-keyboard';
                if (deviceTypeStr === 'mouse') typeIcon = 'fas fa-mouse';

                let isPreferred = '';
                if ((deviceTypeStr === 'keyboard' && device.Name === preferredKeyboard) || 
                    (deviceTypeStr === 'mouse' && device.Name === preferredMouse)) {
                    isPreferred = '<span class="ml-2 text-sm bg-green-100 text-green-800 py-0.5 px-2 rounded">優先デバイス</span>';
                }

                deviceElement.innerHTML = `
                    <div class="flex justify-between items-center">
                        <div>
                            <i class="${typeIcon} mr-2"></i>
                            <strong>${device.Name}</strong> ${isPreferred}
                        </div>
                        <span class="text-xs text-gray-500">${deviceTypeStr}</span>
                    </div>
                    <div class="text-sm text-gray-600 mt-1">${device.Path}</div>
                `;

                // アニメーション付きで追加
                deviceElement.style.opacity = '0';
                deviceElement.style.transform = 'translateY(10px)';
                devicesListDiv.appendChild(deviceElement);

                // フェードイン効果
                setTimeout(() => {
                    deviceElement.style.transition = 'all 0.3s ease';
                    deviceElement.style.opacity = '1';
                    deviceElement.style.transform = 'translateY(0)';
                }, 50 * devices.indexOf(device)); // 連続的に表示

                // セレクトボックスにも追加
                const option = document.createElement('option');
                option.value = device.Name;
                option.textContent = `${device.Name} (${device.Path})`;

                if (deviceTypeStr === 'keyboard') {
                    option.selected = (device.Name === preferredKeyboard);
                    keyboardDeviceSelect.appendChild(option);
                } else if (deviceTypeStr === 'mouse') {
                    option.selected = (device.Name === preferredMouse);
                    mouseDeviceSelect.appendChild(option);
                }
            });

            showToast(`${devices.length}個のデバイスを検出しました`, 'info');
        } catch (error) {
            console.error("デバイス情報取得エラー:", error);
            devicesListDiv.innerHTML = `<p class="text-red-500 text-center">デバイス情報の取得に失敗しました</p>`;
            keyboardDeviceSelect.innerHTML = '<option value="">読み込みエラー</option>';
            mouseDeviceSelect.innerHTML = '<option value="">読み込みエラー</option>';
        }
    }
    getDevicesBtn.addEventListener('click', getDevices);

    setPreferredDevicesBtn.addEventListener('click', async () => {
        const selectedKeyboard = keyboardDeviceSelect.value;
        const selectedMouse = mouseDeviceSelect.value;

        if (!selectedKeyboard && !selectedMouse) {
            showToast('キーボードまたはマウスを選択してください', 'error');
            return;
        }

        try {
            const data = await apiRequest('/devices/preferred', 'PUT', {
                keyboard_device: selectedKeyboard,
                mouse_device: selectedMouse
            });
            showToast('優先デバイスを設定しました', 'success');
            getConfig();
            getDevices();
        } catch (error) {}
    });

    // --- Health Check ---
    async function checkHealth() {
        try {
            const data = await apiRequest('/health');
            updateHealthIndicators(data.status);
            
            // 現在時刻を表示
            lastCheckTime = new Date();
            lastCheckedSpan.textContent = formatDateTime(lastCheckTime);
            
            showToast('システムは正常に動作しています', 'success');
        } catch (error) {
            updateHealthIndicators('error');
            lastCheckedSpan.textContent = formatDateTime(new Date()) + ' (エラー)';
        }
    }
    healthCheckBtn.addEventListener('click', checkHealth);

    // --- 初期化読み込み ---
    async function initializeApp() {
        try {
            // ページ読み込み効果 - 重要な操作を順番に実行
            await getServiceStatus();
            await new Promise(resolve => setTimeout(resolve, 300)); // 視覚的な間隔
            await getConfig();
            await new Promise(resolve => setTimeout(resolve, 200)); // 視覚的な間隔
            await getDevices();
            await new Promise(resolve => setTimeout(resolve, 200)); // 視覚的な間隔
            await checkHealth();
        } catch (error) {
            console.error('アプリの初期化中にエラーが発生しました:', error);
            showToast('アプリの初期化に失敗しました。ネットワークまたはサーバーの状態を確認してください。', 'error');
        }
    }

    // 画面サイズに応じたレスポンシブな調整
    function adjustLayoutForScreenSize() {
        const isMobile = window.innerWidth < 768;
        if (isMobile) {
            document.body.classList.add('mobile-view');
        } else {
            document.body.classList.remove('mobile-view');
        }
    }
    
    window.addEventListener('resize', adjustLayoutForScreenSize);
    adjustLayoutForScreenSize(); // 初期表示時にも実行

    initializeApp();

    // 定期的なステータスチェック
    setInterval(getServiceStatus, 10000);
});
