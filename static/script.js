async function convertHTML() {
  const htmlInput = document.getElementById('htmlInput').value;
  const packagePrefix = document.getElementById('packagePrefix').value;
  const goOutput = document.getElementById('goOutput');

  try {
    const response = await fetch('/convert', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        html: htmlInput,
        packagePrefix: packagePrefix
      }),
    });

    if (!response.ok) {
      throw new Error('Network response was not ok');
    }

    const data = await response.json();
    goOutput.textContent = data.code;
  } catch (error) {
    console.error('Error:', error);
    goOutput.textContent = 'Error converting HTML: ' + error.message;
  }
} 