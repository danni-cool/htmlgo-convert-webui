// 初始化Monaco编辑器
require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.36.1/min/vs' } });

// 编辑器实例
let leftEditor, rightEditor;

// 转换方向和选项
let currentDirection = "html2go"; // 默认方向：HTML到Go
let packagePrefix = "h"; // 默认包前缀
let removePackage = true; // 默认删除package声明

// 定义One Dark Pro主题
const oneDarkPro = {
  base: 'vs-dark',
  inherit: true,
  rules: [
    { token: 'comment', foreground: '5c6370', fontStyle: 'italic' },
    { token: 'keyword', foreground: 'c678dd' },
    { token: 'string', foreground: '98c379' },
    { token: 'number', foreground: 'd19a66' },
    { token: 'type', foreground: '61afef' },
    { token: 'function', foreground: '61afef' },
    { token: 'variable', foreground: 'e06c75' },
    { token: 'constant', foreground: 'd19a66' },
    { token: 'error', foreground: 'e06c75', fontStyle: 'bold underline' },
  ],
  colors: {
    'editor.background': '#282c34',
    'editor.foreground': '#abb2bf',
    'editor.lineHighlightBackground': '#2c313c',
    'editorCursor.foreground': '#528bff',
    'editorWhitespace.foreground': '#3b4048',
    'editorIndentGuide.background': '#3b4048',
    'editor.selectionBackground': '#3e4451',
    'editor.inactiveSelectionBackground': '#3e4451',
    'editorError.foreground': '#e06c75',
    'editorWarning.foreground': '#d19a66',
    'editorInfo.foreground': '#61afef',
  },
};

// 初始化编辑器
require(['vs/editor/editor.main'], function () {
  // 注册One Dark Pro主题
  monaco.editor.defineTheme('oneDarkPro', oneDarkPro);

  // 创建左侧编辑器（默认为HTML输入）
  leftEditor = monaco.editor.create(document.getElementById('leftEditor'), {
    value: '<div class="container">\n  <h1 class="text-xl font-bold">Hello World</h1>\n  <p class="text-gray-600">这是一个示例</p>\n</div>',
    language: 'html',
    theme: 'vs-light',
    minimap: { enabled: false },
    automaticLayout: true,
    fontSize: 14,
    lineHeight: 21,
    padding: { top: 16, bottom: 16 },
    formatOnPaste: true,
    formatOnType: true
  });

  // 创建右侧编辑器（默认为Go输出）
  rightEditor = monaco.editor.create(document.getElementById('rightEditor'), {
    value: '// Go代码将在这里显示',
    language: 'go',
    theme: 'oneDarkPro',
    readOnly: true,
    automaticLayout: true,
    formatOnPaste: true,
    formatOnType: true,
    wordWrap: 'on',
    lineNumbers: 'on',
    renderWhitespace: 'selection',
    scrollBeyondLastLine: false,
    fontSize: 14,
    lineHeight: 21,
    tabSize: 2,
    padding: { top: 16, bottom: 16 },
    renderIndentGuides: true,
    bracketPairColorization: { enabled: true },
    guides: {
      bracketPairs: true,
      indentation: true,
    }
  });

  // 添加左侧编辑器的更改事件监听器
  leftEditor.onDidChangeModelContent(debounce(convertCode, 500));

  // 配置Monaco编辑器的语法校验
  configureEditorValidation();

  // 初始化包前缀输入框
  const pkgInput = document.getElementById('packagePrefix');
  if (pkgInput) {
    pkgInput.value = packagePrefix;
    pkgInput.addEventListener('change', function () {
      packagePrefix = this.value;
      convertCode();
    });
  }

  // 初始化删除package选项
  const removePackageCheckbox = document.getElementById('removePackage');
  if (removePackageCheckbox) {
    removePackageCheckbox.checked = removePackage;
    removePackageCheckbox.addEventListener('change', function () {
      removePackage = this.checked;
      convertCode();
    });
  }

  // 初始化转换按钮
  const convertBtn = document.getElementById('convertBtn');
  if (convertBtn) {
    convertBtn.addEventListener('click', convertCode);
  }

  // 初始化方向切换按钮
  const html2goBtn = document.getElementById('html2goBtn');
  const go2htmlBtn = document.getElementById('go2htmlBtn');

  if (html2goBtn && go2htmlBtn) {
    html2goBtn.addEventListener('click', function () {
      if (currentDirection !== "html2go") {
        switchDirection("html2go");
      }
    });

    go2htmlBtn.addEventListener('click', function () {
      if (currentDirection !== "go2html") {
        switchDirection("go2html");
      }
    });
  }

  // 初始转换
  convertCode();
});

// 配置编辑器语法校验
function configureEditorValidation() {
  // 由于monaco.languages.html.htmlDefaults.setDiagnosticsOptions不可用，
  // 我们使用自定义的验证方法

  // 为左侧编辑器添加自定义校验
  const leftModel = leftEditor.getModel();
  if (leftModel) {
    // 根据当前编辑模式设置校验
    if (currentDirection === "html2go") {
      // HTML校验
      monaco.editor.setModelMarkers(leftModel, 'html', []);
      validateHTML(leftModel);
    } else {
      // Go校验
      monaco.editor.setModelMarkers(leftModel, 'go', []);
      validateGo(leftModel);
    }
  }

  // 添加编辑器内容变化监听器，实时更新语法校验
  leftEditor.onDidChangeModelContent(debounce(function () {
    const model = leftEditor.getModel();
    if (model) {
      if (currentDirection === "html2go") {
        validateHTML(model);
      } else {
        validateGo(model);
      }
    }
  }, 300));
}

// HTML语法校验
function validateHTML(model) {
  const content = model.getValue();
  const markers = [];

  // 检查未闭合的标签
  const openTags = [];
  const tagRegex = /<\/?([a-zA-Z][a-zA-Z0-9]*)(?: [^>]*)?(?:\/)?>/g;
  let match;

  while ((match = tagRegex.exec(content)) !== null) {
    const fullTag = match[0];
    const tagName = match[1];
    const position = model.getPositionAt(match.index);

    // 自闭合标签不需要检查
    if (fullTag.endsWith('/>')) {
      continue;
    }

    // 开始标签
    if (!fullTag.startsWith('</')) {
      openTags.push({ name: tagName, position });
    }
    // 结束标签
    else {
      if (openTags.length === 0) {
        // 多余的结束标签
        markers.push({
          severity: monaco.MarkerSeverity.Error,
          message: `多余的结束标签 </\${tagName}>`,
          startLineNumber: position.lineNumber,
          startColumn: position.column,
          endLineNumber: position.lineNumber,
          endColumn: position.column + fullTag.length
        });
      } else {
        const lastOpenTag = openTags.pop();
        if (lastOpenTag.name !== tagName) {
          // 标签不匹配
          markers.push({
            severity: monaco.MarkerSeverity.Error,
            message: `标签不匹配: 期望 </\${lastOpenTag.name}>, 但找到 </\${tagName}>`,
            startLineNumber: position.lineNumber,
            startColumn: position.column,
            endLineNumber: position.lineNumber,
            endColumn: position.column + fullTag.length
          });
          // 将上一个标签放回，因为它仍然需要被关闭
          openTags.push(lastOpenTag);
        }
      }
    }
  }

  // 检查未闭合的标签
  for (const tag of openTags) {
    markers.push({
      severity: monaco.MarkerSeverity.Error,
      message: `未闭合的标签 <\${tag.name}>`,
      startLineNumber: tag.position.lineNumber,
      startColumn: tag.position.column,
      endLineNumber: tag.position.lineNumber,
      endColumn: tag.position.column + tag.name.length + 1
    });
  }

  // 设置标记
  monaco.editor.setModelMarkers(model, 'html', markers);
}

// Go语法校验
function validateGo(model) {
  const content = model.getValue();
  const markers = [];

  // 检查基本语法错误
  // 1. 检查括号匹配
  const brackets = { '(': ')', '{': '}', '[': ']' };
  const stack = [];

  for (let i = 0; i < content.length; i++) {
    const char = content[i];
    const position = model.getPositionAt(i);

    if (Object.keys(brackets).includes(char)) {
      stack.push({ char, position });
    } else if (Object.values(brackets).includes(char)) {
      if (stack.length === 0) {
        // 多余的右括号
        markers.push({
          severity: monaco.MarkerSeverity.Error,
          message: `多余的右括号 '\${char}'`,
          startLineNumber: position.lineNumber,
          startColumn: position.column,
          endLineNumber: position.lineNumber,
          endColumn: position.column + 1
        });
      } else {
        const last = stack.pop();
        if (brackets[last.char] !== char) {
          // 括号不匹配
          markers.push({
            severity: monaco.MarkerSeverity.Error,
            message: `括号不匹配: 期望 '\${brackets[last.char]}', 但找到 '\${char}'`,
            startLineNumber: position.lineNumber,
            startColumn: position.column,
            endLineNumber: position.lineNumber,
            endColumn: position.column + 1
          });
        }
      }
    }
  }

  // 检查未闭合的括号
  for (const item of stack) {
    markers.push({
      severity: monaco.MarkerSeverity.Error,
      message: `未闭合的左括号 '\${item.char}'`,
      startLineNumber: item.position.lineNumber,
      startColumn: item.position.column,
      endLineNumber: item.position.lineNumber,
      endColumn: item.position.column + 1
    });
  }

  // 设置标记
  monaco.editor.setModelMarkers(model, 'go', markers);
}

// 切换转换方向
function switchDirection(direction) {
  // 保存当前内容
  const currentLeftContent = leftEditor.getValue();
  const currentRightContent = rightEditor.getValue();

  // 更新当前方向
  currentDirection = direction;

  // 更新UI
  if (direction === "html2go") {
    // 切换到HTML到Go方向
    document.getElementById('html2goBtn').classList.remove('bg-white', 'text-gray-900');
    document.getElementById('html2goBtn').classList.add('bg-blue-600', 'text-white');
    document.getElementById('go2htmlBtn').classList.remove('bg-blue-600', 'text-white');
    document.getElementById('go2htmlBtn').classList.add('bg-white', 'text-gray-900');

    document.getElementById('leftEditorTitle').textContent = 'HTML 输入';
    document.getElementById('rightEditorTitle').textContent = 'Go 代码输出';

    document.getElementById('packagePrefixContainer').style.display = 'block';
    document.getElementById('removePackageContainer').style.display = 'flex';

    document.getElementById('htmlExamples').classList.remove('hidden');
    document.getElementById('goExamples').classList.add('hidden');

    // 更改编辑器语言
    monaco.editor.setModelLanguage(leftEditor.getModel(), 'html');
    monaco.editor.setModelLanguage(rightEditor.getModel(), 'go');

    // 更改编辑器主题
    leftEditor.updateOptions({ theme: 'vs-light' });
    rightEditor.updateOptions({ theme: 'oneDarkPro' });

    // 设置编辑器只读状态
    leftEditor.updateOptions({ readOnly: false });
    rightEditor.updateOptions({ readOnly: true });

    // 如果之前是Go到HTML方向，则将右侧的HTML转为左侧的HTML输入
    if (currentRightContent && currentRightContent.trim() !== '<!-- 转换失败 -->' &&
      !currentRightContent.includes('转换错误')) {
      leftEditor.setValue(currentRightContent);
    }
  } else {
    // 切换到Go到HTML方向
    document.getElementById('go2htmlBtn').classList.remove('bg-white', 'text-gray-900');
    document.getElementById('go2htmlBtn').classList.add('bg-blue-600', 'text-white');
    document.getElementById('html2goBtn').classList.remove('bg-blue-600', 'text-white');
    document.getElementById('html2goBtn').classList.add('bg-white', 'text-gray-900');

    document.getElementById('leftEditorTitle').textContent = 'Go 代码输入';
    document.getElementById('rightEditorTitle').textContent = 'HTML 输出';

    document.getElementById('packagePrefixContainer').style.display = 'none';
    document.getElementById('removePackageContainer').style.display = 'none';

    document.getElementById('htmlExamples').classList.add('hidden');
    document.getElementById('goExamples').classList.remove('hidden');

    // 更改编辑器语言
    monaco.editor.setModelLanguage(leftEditor.getModel(), 'go');
    monaco.editor.setModelLanguage(rightEditor.getModel(), 'html');

    // 更改编辑器主题
    leftEditor.updateOptions({ theme: 'oneDarkPro' });
    rightEditor.updateOptions({ theme: 'vs-light' });

    // 设置编辑器只读状态
    leftEditor.updateOptions({ readOnly: false });
    rightEditor.updateOptions({ readOnly: true });

    // 如果之前是HTML到Go方向，则将右侧的Go代码转为左侧的Go输入
    if (currentRightContent && currentRightContent.trim() !== '// Go代码将在这里显示' &&
      currentRightContent.trim() !== '// 转换失败' &&
      !currentRightContent.includes('转换错误')) {
      leftEditor.setValue(currentRightContent);
    }
  }

  // 更新语法校验
  configureEditorValidation();

  // 触发转换
  convertCode();
}

// 防抖函数
function debounce(func, wait) {
  let timeout;
  return function () {
    const context = this;
    const args = arguments;
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(context, args), wait);
  };
}

// 代码转换函数
async function convertCode() {
  try {
    // 首先进行语法校验
    const leftModel = leftEditor.getModel();
    if (leftModel) {
      // 根据当前方向进行校验
      if (currentDirection === "html2go") {
        validateHTML(leftModel);
      } else {
        validateGo(leftModel);
      }

      // 获取当前的标记
      const markers = monaco.editor.getModelMarkers({ resource: leftModel.uri });

      // 如果有错误，显示警告但仍然尝试转换
      if (markers.some(marker => marker.severity === monaco.MarkerSeverity.Error)) {
        console.warn('编辑器中存在语法错误，转换结果可能不正确');
      }
    }

    let requestBody = {};

    if (currentDirection === "html2go") {
      // HTML到Go转换
      const htmlInput = leftEditor.getValue();

      requestBody = {
        html: htmlInput,
        packagePrefix: packagePrefix,
        removePackage: removePackage,
        direction: "html2go"
      };
    } else {
      // Go到HTML转换
      const goInput = leftEditor.getValue();

      requestBody = {
        goCode: goInput,
        direction: "go2html"
      };
    }

    const response = await fetch('/convert', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    });

    if (!response.ok) {
      // 尝试解析错误响应为JSON
      let errorData;
      try {
        errorData = await response.json();
      } catch (e) {
        // 如果不是JSON，则获取文本
        const errorText = await response.text();
        throw new Error(errorText);
      }

      // 如果成功解析为JSON，则使用其中的错误信息
      if (errorData && errorData.error) {
        throw new Error(errorData.error);
      } else {
        throw new Error('转换失败: 未知错误');
      }
    }

    const data = await response.json();

    // 根据当前方向设置输出
    if (currentDirection === "html2go") {
      rightEditor.setValue(data.code || '// 转换失败');

      // 检查转换后的Go代码是否有语法错误
      if (data.code) {
        const rightModel = rightEditor.getModel();
        if (rightModel) {
          // 清除之前的标记
          monaco.editor.setModelMarkers(rightModel, 'go', []);

          // 检查生成的Go代码
          setTimeout(() => {
            // 使用setTimeout确保编辑器已经更新
            const goMarkers = [];

            // 检查常见的Go语法错误
            if (data.code.includes('undefined:')) {
              const lines = data.code.split('\n');
              for (let i = 0; i < lines.length; i++) {
                if (lines[i].includes('undefined:')) {
                  goMarkers.push({
                    severity: monaco.MarkerSeverity.Error,
                    message: `未定义的标识符: ${lines[i]}`,
                    startLineNumber: i + 1,
                    startColumn: 1,
                    endLineNumber: i + 1,
                    endColumn: lines[i].length + 1
                  });
                }
              }
            }

            // 设置标记
            monaco.editor.setModelMarkers(rightModel, 'go', goMarkers);
          }, 100);
        }
      }
    } else {
      rightEditor.setValue(data.html || '<!-- 转换失败 -->');

      // 检查转换后的HTML是否有语法错误
      if (data.html) {
        const rightModel = rightEditor.getModel();
        if (rightModel) {
          // 清除之前的标记
          monaco.editor.setModelMarkers(rightModel, 'html', []);

          // 检查生成的HTML代码
          setTimeout(() => {
            // 使用setTimeout确保编辑器已经更新
            validateHTML(rightModel);
          }, 100);
        }
      }
    }
  } catch (error) {
    console.error('Error:', error);

    // 显示错误信息
    if (currentDirection === "html2go") {
      rightEditor.setValue('// 转换错误: ' + error.message);
    } else {
      rightEditor.setValue('<!-- 转换错误: ' + error.message + ' -->');
    }

    // 在编辑器中标记错误位置（如果可能）
    const leftModel = leftEditor.getModel();
    if (leftModel) {
      const errorMessage = error.message;
      const markers = [];

      // 尝试从错误消息中提取行号和列号
      const lineMatch = errorMessage.match(/line (\d+)/i);
      const colMatch = errorMessage.match(/column (\d+)/i);

      if (lineMatch && lineMatch[1]) {
        const lineNumber = parseInt(lineMatch[1], 10);
        const columnNumber = colMatch && colMatch[1] ? parseInt(colMatch[1], 10) : 1;
        const lineContent = leftModel.getLineContent(lineNumber);

        markers.push({
          severity: monaco.MarkerSeverity.Error,
          message: errorMessage,
          startLineNumber: lineNumber,
          startColumn: columnNumber,
          endLineNumber: lineNumber,
          endColumn: lineContent.length + 1
        });
      } else {
        // 尝试从错误消息中提取标签名或变量名
        const tagMatch = errorMessage.match(/<([a-zA-Z][a-zA-Z0-9]*)>/);
        const varMatch = errorMessage.match(/undefined:\s*([a-zA-Z][a-zA-Z0-9]*)/);

        if (tagMatch && tagMatch[1]) {
          // 搜索HTML中的标签
          const tagName = tagMatch[1];
          const content = leftModel.getValue();
          const tagRegex = new RegExp(`<${tagName}[^>]*>`, 'g');
          let match;

          while ((match = tagRegex.exec(content)) !== null) {
            const position = leftModel.getPositionAt(match.index);
            markers.push({
              severity: monaco.MarkerSeverity.Error,
              message: errorMessage,
              startLineNumber: position.lineNumber,
              startColumn: position.column,
              endLineNumber: position.lineNumber,
              endColumn: position.column + match[0].length
            });
          }
        } else if (varMatch && varMatch[1]) {
          // 搜索Go代码中的变量
          const varName = varMatch[1];
          const content = leftModel.getValue();
          const varRegex = new RegExp(`\\b${varName}\\b`, 'g');
          let match;

          while ((match = varRegex.exec(content)) !== null) {
            const position = leftModel.getPositionAt(match.index);
            markers.push({
              severity: monaco.MarkerSeverity.Error,
              message: errorMessage,
              startLineNumber: position.lineNumber,
              startColumn: position.column,
              endLineNumber: position.lineNumber,
              endColumn: position.column + varName.length
            });
          }
        }

        // 如果没有找到具体位置，标记整个文档
        if (markers.length === 0) {
          markers.push({
            severity: monaco.MarkerSeverity.Error,
            message: errorMessage,
            startLineNumber: 1,
            startColumn: 1,
            endLineNumber: leftModel.getLineCount(),
            endColumn: 1
          });
        }
      }

      // 设置标记
      monaco.editor.setModelMarkers(leftModel, currentDirection === "html2go" ? 'html' : 'go', markers);
    }
  }
}

// HTML示例代码
function loadHTMLExample(id) {
  let exampleHTML = '';

  switch (id) {
    case 1:
      exampleHTML = `<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold mb-4">Hello World</h1>
  <p class="text-gray-600">这是一个基本的HTML示例</p>
  <button class="bg-blue-500 text-white px-4 py-2 rounded mt-4">
    点击我
  </button>
</div>`;
      break;
    case 2:
      exampleHTML = `<form class="max-w-md mx-auto p-6 bg-white rounded-lg shadow-md">
  <div class="mb-4">
    <label class="block text-gray-700 text-sm font-bold mb-2" for="username">
      用户名
    </label>
    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline" id="username" type="text" placeholder="用户名" required>
  </div>
  <div class="mb-6">
    <label class="block text-gray-700 text-sm font-bold mb-2" for="password">
      密码
    </label>
    <input class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 mb-3 leading-tight focus:outline-none focus:shadow-outline" id="password" type="password" placeholder="******************" required>
  </div>
  <div class="flex items-center justify-between">
    <button class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline" type="submit">
      登录
    </button>
    <a class="inline-block align-baseline font-bold text-sm text-blue-500 hover:text-blue-800" href="#">
      忘记密码?
    </a>
  </div>
</form>`;
      break;
    case 3:
      exampleHTML = `<div class="max-w-6xl mx-auto p-4">
  <header class="bg-white shadow rounded-lg p-4 mb-6">
    <div class="flex justify-between items-center">
      <div class="flex items-center">
        <img src="https://via.placeholder.com/50" alt="Logo" class="h-10 w-10 mr-3">
        <h1 class="text-xl font-bold text-gray-800">我的应用</h1>
      </div>
      <nav>
        <ul class="flex space-x-4">
          <li><a href="#" class="text-blue-600 hover:text-blue-800">首页</a></li>
          <li><a href="#" class="text-gray-600 hover:text-gray-800">关于</a></li>
          <li><a href="#" class="text-gray-600 hover:text-gray-800">服务</a></li>
          <li><a href="#" class="text-gray-600 hover:text-gray-800">联系我们</a></li>
        </ul>
      </nav>
    </div>
  </header>
  
  <main class="grid grid-cols-1 md:grid-cols-3 gap-6">
    <aside class="md:col-span-1 bg-white p-4 rounded-lg shadow">
      <h2 class="text-lg font-semibold mb-4">侧边栏</h2>
      <ul class="space-y-2">
        <li><a href="#" class="block p-2 bg-gray-100 rounded hover:bg-gray-200">菜单项 1</a></li>
        <li><a href="#" class="block p-2 rounded hover:bg-gray-200">菜单项 2</a></li>
        <li><a href="#" class="block p-2 rounded hover:bg-gray-200">菜单项 3</a></li>
      </ul>
    </aside>
    
    <section class="md:col-span-2 bg-white p-4 rounded-lg shadow">
      <h2 class="text-xl font-bold mb-4">主要内容</h2>
      <p class="text-gray-700 mb-4">这是一个复杂布局示例，展示了如何使用HTML和Tailwind CSS创建响应式布局。</p>
      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
        <div class="bg-gray-100 p-4 rounded">
          <h3 class="font-semibold mb-2">卡片 1</h3>
          <p class="text-sm">这是卡片内容的描述文本。</p>
        </div>
        <div class="bg-gray-100 p-4 rounded">
          <h3 class="font-semibold mb-2">卡片 2</h3>
          <p class="text-sm">这是卡片内容的描述文本。</p>
        </div>
      </div>
      <button class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">了解更多</button>
    </section>
  </main>
</div>`;
      break;
  }

  // 确保当前是HTML到Go方向
  if (currentDirection !== "html2go") {
    switchDirection("html2go");
  }

  if (leftEditor) {
    leftEditor.setValue(exampleHTML);
  }
}

// Go示例代码
function loadGoExample(id) {
  let exampleGo = '';

  switch (id) {
    case 1:
      exampleGo = `var n = htmlgo.Div(htmlgo.H1("Hello World").Class("text-2xl font-bold mb-4"), htmlgo.P(htmlgo.Text("这是一个基本的Go示例")).Class("text-gray-600"), htmlgo.Button("点击我").Class("bg-blue-500 text-white px-4 py-2 rounded mt-4")).Class("container mx-auto p-4")`;
      break;
    case 2:
      exampleGo = `var n = htmlgo.Form(htmlgo.Div(htmlgo.Label("用户名").Class("block text-gray-700 text-sm font-bold mb-2").Attr("for", "username"), htmlgo.Input("username").Type("text").Attr("placeholder", "用户名").Attr("required", "true").Class("shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline")).Class("mb-4"), htmlgo.Div(htmlgo.Label("密码").Class("block text-gray-700 text-sm font-bold mb-2").Attr("for", "password"), htmlgo.Input("password").Type("password").Attr("placeholder", "******************").Attr("required", "true").Class("shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 mb-3 leading-tight focus:outline-none focus:shadow-outline")).Class("mb-6"), htmlgo.Div(htmlgo.Button("登录").Type("submit").Class("bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"), htmlgo.A(htmlgo.Text("忘记密码?")).Attr("href", "#").Class("inline-block align-baseline font-bold text-sm text-blue-500 hover:text-blue-800")).Class("flex items-center justify-between")).Class("max-w-md mx-auto p-6 bg-white rounded-lg shadow-md")`;
      break;
    case 3:
      exampleGo = `var n = htmlgo.Div(htmlgo.Header(htmlgo.Div(htmlgo.Text("导航栏")).Class("flex justify-between items-center")).Class("bg-white shadow rounded-lg p-4 mb-6"), htmlgo.Main(htmlgo.Aside(htmlgo.H2("侧边栏").Class("text-lg font-semibold mb-4")).Class("md:col-span-1 bg-white p-4 rounded-lg shadow"), htmlgo.Section(htmlgo.H2("主要内容").Class("text-xl font-bold mb-4"), htmlgo.P(htmlgo.Text("这是一个复杂布局示例")).Class("text-gray-700 mb-4")).Class("md:col-span-2 bg-white p-4 rounded-lg shadow")).Class("grid grid-cols-1 md:grid-cols-3 gap-6")).Class("max-w-6xl mx-auto p-4")`;
      break;
  }

  // 确保当前是Go到HTML方向
  if (currentDirection !== "go2html") {
    switchDirection("go2html");
  }

  if (leftEditor) {
    leftEditor.setValue(exampleGo);
  }
}

// 添加编辑器内容变化监听器
function setupEditorListeners() {
  // 左侧编辑器内容变化时触发代码转换
  // 注意：语法校验的监听器已经在configureEditorValidation中添加
  leftEditor.onDidChangeModelContent(debounce(function () {
    // 触发代码转换
    convertCode();
  }, 500));
} 