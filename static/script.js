// 初始化Monaco编辑器
require.config({ paths: { vs: 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.36.1/min/vs' } });

// 编辑器实例
let htmlEditor, goEditor;

// 转换选项
let packagePrefix = "h"; // 默认包前缀
let isUpdating = false; // 防止无限循环更新的标志

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

  // 安全地禁用HTML和JavaScript语法校验
  try {
    if (monaco.languages.html && monaco.languages.html.htmlDefaults &&
      typeof monaco.languages.html.htmlDefaults.setDiagnosticsOptions === 'function') {
      monaco.languages.html.htmlDefaults.setDiagnosticsOptions({
        validate: false
      });
    }
  } catch (e) {
    console.warn('无法配置HTML语言校验', e);
  }

  try {
    if (monaco.languages.typescript && monaco.languages.typescript.javascriptDefaults &&
      typeof monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions === 'function') {
      monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
        noSemanticValidation: true,
        noSyntaxValidation: true
      });
    }
  } catch (e) {
    console.warn('无法配置JavaScript语言校验', e);
  }

  // 尝试禁用Go语言的语法校验（如果Monaco支持）
  try {
    if (monaco.languages.go && monaco.languages.go.goDefaults &&
      typeof monaco.languages.go.goDefaults.setDiagnosticsOptions === 'function') {
      monaco.languages.go.goDefaults.setDiagnosticsOptions({
        validate: false,
        noSemanticValidation: true,
        noSyntaxValidation: true
      });
    }
  } catch (e) {
    console.warn('无法配置Go语言校验', e);
  }

  // 创建HTML编辑器
  htmlEditor = monaco.editor.create(document.getElementById('leftEditor'), {
    value: '<div class="container">\n  <h1 class="text-xl font-bold">Hello World</h1>\n  <p class="text-gray-600">这是一个示例</p>\n</div>',
    language: 'html',
    theme: 'vs-light',
    minimap: { enabled: false },
    automaticLayout: true,
    fontSize: 14,
    lineHeight: 21,
    padding: { top: 16, bottom: 16 },
    formatOnPaste: true,
    formatOnType: true,
    // 禁用语法校验
    validate: false
  });

  // 创建Go编辑器
  createGoEditor();

  // 配置编辑器的语法校验
  configureEditorValidation();

  // 设置编辑器的事件监听器
  setupEditorListeners();

  // 初始化包前缀输入框
  const pkgInput = document.getElementById('packagePrefix');
  if (pkgInput) {
    pkgInput.value = packagePrefix;
    pkgInput.addEventListener('change', function () {
      // 更新包前缀变量，即使是空值
      packagePrefix = this.value;
      // 如果当前HTML编辑器有内容，则触发HTML到Go的转换
      if (htmlEditor.getValue().trim() !== '') {
        htmlToGoConversion();
      }
    });
  }

  // 清除所有编辑器标记
  if (htmlEditor && htmlEditor.getModel()) {
    monaco.editor.setModelMarkers(htmlEditor.getModel(), 'html', []);
  }
  if (goEditor && goEditor.getModel()) {
    monaco.editor.setModelMarkers(goEditor.getModel(), 'go', []);
  }

  // 初始转换
  htmlToGoConversion();
});

// 创建Go编辑器的函数
function createGoEditor() {
  // 如果已经存在编辑器，先销毁它
  if (goEditor) {
    goEditor.dispose();
  }

  // 创建新的Go编辑器
  goEditor = monaco.editor.create(document.getElementById('rightEditor'), {
    value: '// Go代码将在这里显示',
    language: 'go',
    theme: 'oneDarkPro',
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
    },
    readOnly: false, // 确保编辑器不是只读的
    domReadOnly: false, // 确保DOM元素不是只读的
    // 禁用语法校验
    validate: false
  });

  console.log("Go编辑器已重新创建");

  // 添加点击事件，确保编辑器可以获得焦点
  document.getElementById('rightEditor').addEventListener('click', function () {
    if (goEditor) {
      goEditor.focus();
      console.log("Go编辑器已获得焦点");
    }
  });
}

// 配置编辑器语法校验
function configureEditorValidation() {
  // 禁用所有语法校验
  const htmlModel = htmlEditor.getModel();
  if (htmlModel) {
    monaco.editor.setModelMarkers(htmlModel, 'html', []);
    // validateHTML(htmlModel); // 已禁用
  }

  // Go校验
  const goModel = goEditor.getModel();
  if (goModel) {
    monaco.editor.setModelMarkers(goModel, 'go', []);
    // validateGo(goModel); // 已禁用
  }

  // 禁用编辑器内容变化监听器的语法校验
  // htmlEditor.onDidChangeModelContent(debounce(function () {
  //   const model = htmlEditor.getModel();
  //   if (model) {
  //     validateHTML(model);
  //   }
  // }, 300));

  // goEditor.onDidChangeModelContent(debounce(function () {
  //   const model = goEditor.getModel();
  //   if (model) {
  //     validateGo(model);
  //   }
  // }, 300));
}

// HTML语法校验 - 已禁用
function validateHTML(model) {
  // 函数保留但不执行任何校验
  return;

  // 以下代码已被禁用
  // const content = model.getValue();
  // const markers = [];
  // ...
}

// Go语法校验 - 已禁用
function validateGo(model) {
  // 函数保留但不执行任何校验
  return;

  // 以下代码已被禁用
  // const content = model.getValue();
  // const markers = [];
  // ...
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

// HTML到Go转换函数
async function htmlToGoConversion() {
  if (isUpdating) return;
  isUpdating = true;

  try {
    // 获取HTML内容
    const htmlInput = htmlEditor.getValue();
    console.log("HTML输入内容:", htmlInput);

    if (!htmlInput.trim()) {
      goEditor.setValue('// 请在左侧输入HTML代码');
      isUpdating = false;
      return;
    }

    // 准备请求体
    const requestBody = {
      html: htmlInput,
      packagePrefix: packagePrefix,
      direction: "html2go"
    };

    console.log("发送转换请求:", JSON.stringify(requestBody));

    // 发送转换请求
    const response = await fetch('/convert', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    });

    console.log("收到响应状态:", response.status);

    if (!response.ok) {
      // 尝试解析错误响应
      let errorData;
      try {
        errorData = await response.json();
        console.log("错误响应数据:", errorData);
      } catch (e) {
        const errorText = await response.text();
        console.log("错误响应文本:", errorText);
        throw new Error(errorText);
      }

      if (errorData && errorData.error) {
        throw new Error(errorData.error);
      } else {
        throw new Error('转换失败: 未知错误');
      }
    }

    const data = await response.json();
    console.log("转换成功，响应数据:", data);

    // 更新Go编辑器
    goEditor.setValue(data.code || '// 转换失败');
  } catch (error) {
    console.error('HTML到Go转换错误:', error);
    goEditor.setValue(`// 转换错误: ${error.message}`);
  } finally {
    isUpdating = false;
  }
}

// Go到HTML的转换
function goToHtmlConversion() {
  if (isUpdating) {
    return;
  }

  isUpdating = true;
  console.log("开始Go到HTML转换");
  const goCode = goEditor.getValue();
  console.log("获取到Go代码:", goCode);

  if (!goCode || !goCode.trim()) {
    console.log("Go代码为空，不执行转换");
    htmlEditor.setValue('');
    isUpdating = false;
    return;
  }

  // 检查Go代码是否包含变量n的定义
  let processedGoCode = goCode;

  // 检查直接if表达式错误 (最常见的错误模式)
  if (goCode.match(/var\s+n\s*=\s*if/) || goCode.match(/n\s*:=\s*if/) ||
    goCode.includes('n = if') || goCode.includes('= if true') ||
    goCode.includes('= if false')) {
    // 直接if条件表达式错误
    htmlEditor.setValue(`<!-- 编译或执行错误: syntax error: unexpected if, expected expression -->
<!-- 请尝试以下正确写法: -->
<!--
    // 方法1: 使用函数
    var n = htmlgo.Div().Text(func() string {
        if condition {
            return "真"
        }
        return "假"
    }())
    
    // 方法2: 使用map模拟三元运算符
    condition := true
    var n = htmlgo.Div().Text(map[bool]string{true: "真", false: "假"}[condition])
-->
`);
    isUpdating = false;
    return;
  }

  // 检查常见语法错误
  if (goCode.includes('if') &&
    (goCode.includes('if {') ||
      goCode.includes('if true {') ||
      goCode.includes('if false {') ||
      goCode.match(/var\s+n\s*=\s*if/))) {
    // 条件语句语法错误
    htmlEditor.setValue(`<!-- 警告: Go代码中存在条件语句语法错误，请检查您的代码 -->`);
    isUpdating = false;
    return;
  }

  // 检查括号匹配
  const openBraces = (goCode.match(/\{/g) || []).length;
  const closeBraces = (goCode.match(/\}/g) || []).length;
  if (openBraces !== closeBraces) {
    htmlEditor.setValue(`<!-- 警告: Go代码中的花括号不匹配，开括号: ${openBraces}，闭括号: ${closeBraces} -->`);
    isUpdating = false;
    return;
  }

  const openParens = (goCode.match(/\(/g) || []).length;
  const closeParens = (goCode.match(/\)/g) || []).length;
  if (openParens !== closeParens) {
    htmlEditor.setValue(`<!-- 警告: Go代码中的圆括号不匹配，开括号: ${openParens}，闭括号: ${closeParens} -->`);
    isUpdating = false;
    return;
  }

  if (!goCode.includes('var n =') && !goCode.includes('n :=')) {
    // 尝试提取主要表达式
    const lines = goCode.split('\n');
    let mainExpr = '';

    // 找到第一个非空且非注释的行
    for (const line of lines) {
      const trimmedLine = line.trim();
      if (trimmedLine && !trimmedLine.startsWith('//')) {
        mainExpr = trimmedLine;
        break;
      }
    }

    if (mainExpr) {
      // 将主要表达式包装在变量定义中
      processedGoCode = `var n = ${mainExpr}`;
      console.log("处理后的Go代码:", processedGoCode);
    }
  }

  const requestBody = {
    goCode: processedGoCode,
    direction: "go2html"
  };

  console.log("发送请求:", JSON.stringify(requestBody));

  fetch('/convert', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(requestBody)
  })
    .then(response => {
      console.log("收到响应状态:", response.status);
      if (!response.ok) {
        throw new Error(`HTTP错误! 状态: ${response.status}`);
      }
      return response.json();
    })
    .then(data => {
      console.log("收到响应数据:", data);
      if (data.error) {
        console.error("转换错误:", data.error);
        htmlEditor.setValue(`<!-- 转换错误: ${data.error} -->`);
      } else {
        htmlEditor.setValue(data.html || '');
      }
    })
    .catch(error => {
      console.error('转换错误:', error);
      htmlEditor.setValue(`<!-- 转换错误: ${error.message} -->`);
    })
    .finally(() => {
      isUpdating = false;
      console.log("Go到HTML转换完成");
    });
}

// 加载示例代码
function loadExample(id) {
  let htmlExample = '';

  switch (id) {
    case 1:
      // 基本结构示例
      htmlExample = `<div class="container mx-auto p-4">
  <h1 class="text-2xl font-bold text-blue-600">Hello World</h1>
  <p class="mt-2 text-gray-600">这是一个简单的HTML示例，使用了Tailwind CSS类。</p>
  <button class="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
    点击我
  </button>
</div>`;
      break;
    case 2:
      // 表单示例
      htmlExample = `<form class="max-w-md mx-auto p-6 bg-white rounded-lg shadow-md">
  <h2 class="text-xl font-semibold mb-4">联系表单</h2>
  <div class="mb-4">
    <label class="block text-gray-700 text-sm font-bold mb-2" for="name">
      姓名
    </label>
    <input
      class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      id="name"
      type="text"
      placeholder="请输入您的姓名"
      required
    />
  </div>
  <div class="mb-4">
    <label class="block text-gray-700 text-sm font-bold mb-2" for="email">
      邮箱
    </label>
    <input
      class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      id="email"
      type="email"
      placeholder="请输入您的邮箱"
      required
    />
  </div>
  <div class="mb-6">
    <label class="block text-gray-700 text-sm font-bold mb-2" for="message">
      留言
    </label>
    <textarea
      class="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
      id="message"
      rows="4"
      placeholder="请输入您的留言"
    ></textarea>
  </div>
  <div class="flex items-center justify-between">
    <button
      class="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
      type="submit"
    >
      提交
    </button>
    <button
      class="bg-gray-300 hover:bg-gray-400 text-gray-800 font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
      type="reset"
    >
      重置
    </button>
  </div>
</form>`;
      break;
    case 3:
      // 复杂布局示例
      htmlExample = `<div class="max-w-6xl mx-auto p-4">
  <header class="bg-white shadow rounded-lg p-4 mb-6">
    <div class="flex justify-between items-center">
      <div class="flex items-center">
        <img src="https://via.placeholder.com/50" alt="Logo" class="h-10 w-10 mr-3" />
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
    <aside class="md:col-span-1">
      <div class="bg-white shadow rounded-lg p-4">
        <h2 class="text-lg font-semibold mb-4">侧边栏导航</h2>
        <ul class="space-y-2">
          <li class="p-2 bg-blue-50 rounded text-blue-600">仪表盘</li>
          <li class="p-2 hover:bg-gray-50 rounded">用户管理</li>
          <li class="p-2 hover:bg-gray-50 rounded">产品列表</li>
          <li class="p-2 hover:bg-gray-50 rounded">订单管理</li>
          <li class="p-2 hover:bg-gray-50 rounded">设置</li>
        </ul>
      </div>
    </aside>

    <div class="md:col-span-2">
      <div class="bg-white shadow rounded-lg p-4 mb-6">
        <h2 class="text-lg font-semibold mb-4">欢迎回来</h2>
        <p class="text-gray-600">
          这是一个复杂布局示例，展示了如何使用Tailwind CSS创建响应式布局。
        </p>
        <div class="mt-4">
          <button class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
            开始使用
          </button>
        </div>
      </div>

      <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div class="bg-white shadow rounded-lg p-4">
          <h3 class="font-semibold text-gray-800">统计数据</h3>
          <p class="text-3xl font-bold text-blue-600 mt-2">1,234</p>
          <p class="text-sm text-gray-500">总用户数</p>
        </div>
        <div class="bg-white shadow rounded-lg p-4">
          <h3 class="font-semibold text-gray-800">收入</h3>
          <p class="text-3xl font-bold text-green-600 mt-2">$5,678</p>
          <p class="text-sm text-gray-500">本月收入</p>
        </div>
      </div>
    </div>
  </main>

  <footer class="mt-8 bg-white shadow rounded-lg p-4 text-center text-gray-500">
    <p>© 2023 我的应用. 保留所有权利.</p>
  </footer>
</div>`;
      break;
    case 4:
      // VuetifyX Dialog 基本示例
      htmlExample = `<div>
  <vx-dialog
    title="确认"
    text="这是一个基本的确认对话框示例"
    ok-text="确定"
    cancel-text="取消"
  >
    <v-btn color="primary">打开对话框</v-btn>
  </vx-dialog>
</div>`;
      break;
    case 5:
      // VuetifyX DatePicker 示例
      htmlExample = `<div>
  <vx-date-picker
    label="日期选择器"
    clearable
    tips="示例提示文本"
  />
</div>`;
      break;
    case 6:
      // VuetifyX TiptapEditor 富文本编辑器示例
      htmlExample = `<div>
  <vx-tiptap-editor
    label="富文本编辑器"
    min-height="200"
  />
</div>`;
      break;
    case 7:
      // VuetifyX Dialog 高级示例
      htmlExample = `<div>
  <vx-dialog
    title="确认删除"
    text="您确定要删除这条记录吗？此操作不可撤销。"
    icon="mdi-delete-alert"
    icon-color="error"
    ok-text="删除"
    ok-color="error"
    cancel-text="取消"
  >
    <v-btn color="error">删除记录</v-btn>
  </vx-dialog>
</div>`;
      break;
    default:
      htmlExample = '<div>示例加载失败</div>';
  }

  console.log(`加载示例 ${id}`); // 添加日志，帮助调试

  // 设置HTML编辑器内容
  if (htmlEditor) {
    htmlEditor.setValue(htmlExample);
    // 手动触发HTML到Go转换，增加延迟时间
    setTimeout(() => {
      console.log("准备触发HTML到Go转换...");
      htmlToGoConversion();
      console.log("已触发HTML到Go转换");
    }, 500); // 增加延迟时间到500毫秒
  } else {
    console.error('HTML编辑器未初始化');
  }
}

// 添加编辑器内容变化监听器
function setupEditorListeners() {
  // 不再自动转换，改为手动按钮触发

  // 添加HTML到Go转换按钮事件
  const htmlToGoBtn = document.getElementById('htmlToGoBtn');
  if (htmlToGoBtn) {
    htmlToGoBtn.addEventListener('click', function () {
      console.log("点击了HTML到Go转换按钮");
      htmlToGoConversion();
    });
  }

  // 添加Go到HTML转换按钮事件
  const goToHtmlBtn = document.getElementById('goToHtmlBtn');
  if (goToHtmlBtn) {
    goToHtmlBtn.addEventListener('click', function () {
      console.log("点击了Go到HTML转换按钮");
      goToHtmlConversion();
    });
  }

  // 确保Go编辑器是可编辑的
  setTimeout(function () {
    // 强制设置Go编辑器为可编辑状态
    if (goEditor) {
      goEditor.updateOptions({ readOnly: false, domReadOnly: false });
      console.log("已强制设置Go编辑器为可编辑状态");

      // 获取编辑器DOM元素并确保它不是只读的
      const goEditorElement = document.getElementById('rightEditor');
      if (goEditorElement) {
        const editorDomNode = goEditorElement.querySelector('.monaco-editor');
        if (editorDomNode) {
          editorDomNode.setAttribute('aria-readonly', 'false');
          console.log("已设置编辑器DOM元素为非只读");
        }
      }
    }
  }, 1000);
}