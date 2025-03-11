document.addEventListener('DOMContentLoaded', function() {
  // Предварительно загружаем стили для обеих тем, чтобы избежать задержек при переключении
  function preloadThemeStyles() {
    // Создаем скрытый div для предзагрузки стилей темной темы
    const preloadDiv = document.createElement('div');
    preloadDiv.className = 'dark hidden';
    preloadDiv.style.position = 'absolute';
    preloadDiv.style.opacity = '0';
    preloadDiv.style.pointerEvents = 'none';
    document.body.appendChild(preloadDiv);
    
    // Принудительно вычисляем стили
    window.getComputedStyle(preloadDiv).backgroundColor;
    
    // Удаляем элемент после предзагрузки
    setTimeout(() => {
      document.body.removeChild(preloadDiv);
    }, 100);
  }
  
  // Вызываем функцию предзагрузки стилей
  preloadThemeStyles();
  
  // Get the export button
  const exportButton = document.getElementById('export-pdf');
  
  if (exportButton) {
    exportButton.addEventListener('click', function() {
      // Show a loading indicator or message
      const originalContent = exportButton.innerHTML;
      exportButton.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><path d="M12 6v6l4 2"></path></svg>';
      exportButton.disabled = true;
      
      // Send request to the server to generate PDF
      fetch('/export-pdf')
        .then(response => {
          if (!response.ok) {
            throw new Error('Network response was not ok');
          }
          return response.blob();
        })
        .then(blob => {
          // Create a URL for the blob
          const url = window.URL.createObjectURL(blob);
          
          // Create a temporary link element
          const a = document.createElement('a');
          a.style.display = 'none';
          a.href = url;
          a.download = 'resume.pdf';
          
          // Append to the document and trigger the download
          document.body.appendChild(a);
          a.click();
          
          // Clean up
          window.URL.revokeObjectURL(url);
          document.body.removeChild(a);
          
          // Reset the button
          exportButton.innerHTML = originalContent;
          exportButton.disabled = false;
        })
        .catch(error => {
          console.error('Error exporting PDF:', error);
          
          // Show error message
          alert('Failed to export PDF. Please try again.');
          
          // Reset the button
          exportButton.innerHTML = originalContent;
          exportButton.disabled = false;
        });
    });
  }

  // Check for saved theme preference or use system preference
  function initTheme() {
    const darkMode = localStorage.getItem('darkMode');
    
    if (darkMode === null) {
      // Use system preference
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      localStorage.setItem('darkMode', prefersDark);
      document.documentElement.classList.toggle('dark', prefersDark);
    } else {
      document.documentElement.classList.toggle('dark', darkMode === 'true');
    }
  }

  // Initialize theme
  initTheme();

  // Toggle theme
  const themeToggle = document.querySelector('.theme-toggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', function() {
      // Оптимизация переключения темы
      // Сначала добавляем/удаляем класс
      const isDark = document.documentElement.classList.toggle('dark');
      localStorage.setItem('darkMode', isDark);
      
      // Затем оптимизируем анимацию для списков
      const lists = document.querySelectorAll('ul.list-disc');
      lists.forEach(list => {
        // Временно отключаем переходы для списков
        list.style.transition = 'none';
        
        // Принудительно применяем стили
        window.getComputedStyle(list).opacity;
        
        // Восстанавливаем переходы через небольшую задержку
        setTimeout(() => {
          list.style.transition = '';
        }, 50);
      });
      
      // Специальная обработка для оптимизированных списков
      const optimizedLists = document.querySelectorAll('.theme-optimized-list');
      optimizedLists.forEach(list => {
        // Полностью отключаем анимацию для списков
        list.style.transition = 'none';
      });
      
      // Оптимизируем анимацию для элементов списка
      const listItems = document.querySelectorAll('li');
      listItems.forEach(item => {
        // Применяем цвет напрямую для более быстрого обновления
        item.style.color = isDark ? 
          'hsl(0, 0%, 98%)' : // Светлый текст для темной темы
          'hsl(240, 10%, 3.9%)'; // Темный текст для светлой темы
      });
      
      // Специальная обработка для оптимизированных элементов списка
      const optimizedItems = document.querySelectorAll('.theme-optimized-item');
      optimizedItems.forEach(item => {
        // Применяем цвет с короткой анимацией
        item.style.transition = 'color 0.15s ease';
        item.style.color = isDark ? 
          'hsl(0, 0%, 98%)' : // Светлый текст для темной темы
          'hsl(240, 10%, 3.9%)'; // Темный текст для светлой темы
      });
      
      // Создаем пользовательское событие для обновления других компонентов
      document.dispatchEvent(new CustomEvent('themeChanged', { detail: { isDark } }));
    });
  }
}); 