// ==UserScript==
// @name         Thepd categories
// @namespace    http://tampermonkey.net/
// @version      2025-07-19
// @description  提取 category-container 分类信息并发送到 API，并支持手动触发发送逻辑
// @description  设置: 允许跨域。执行方式: 页面加载执行; 手动点击菜单执行;
// @author       You
// @match        https://theporndude.com/*
// @icon         https://www.google.com/s2/favicons?sz=64&domain=tampermonkey.net
// @grant        GM_registerMenuCommand
// //#grant        GM_xmlhttpRequest
// @grant        GM.xmlhttpRequest
// ==/UserScript==

(function() {
    'use strict';

    console.log('the pd started.');

    // 配置信息
    const CONFIG = {
        apiUrl: 'http://192.168.194.135:8000/api/admin/category',
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjEsImV4cCI6MTc4NDQyNDE1MCwibmJmIjoxNzUyODg4MTUwLCJpYXQiOjE3NTI4ODgxNTB9.daD8-FufO1oH8heJv1ysemi3To3ycHZkHDeGWntSDkI',
        targetClass: 'category-container'
    };

    // 等待页面完全加载
    function waitForPageLoad() {
        return new Promise((resolve) => {
            if (document.readyState === 'complete') {
                resolve();
            } else {
                window.addEventListener('load', resolve);
            }
        });
    }

    // 查找包含指定class的节点
    function findCategoryContainers() {
        const containers = document.querySelectorAll(`.${CONFIG.targetClass}`);
        console.log(`Found ${containers.length} category-container elements`);
        return containers;
    }

    // 从category-container中提取所需数据
    function extractCategoryData(container) {
        try {
            // title: .category-container > .category-header > h2 > a 的text属性
            const titleElement = container.querySelector('.category-header > h2 > a');
            const title = titleElement ? titleElement.textContent.trim() : '';

            // url: .category-container > .category-header > h2 > a 的href属性
            const url = titleElement ? titleElement.href : '';

            // desc: .category-container > .desc 的text属性
            const descElement = container.querySelector('.desc');
            const desc = descElement ? descElement.textContent.trim() : '';

            // iconcss: .category-container > .icon-category 的class属性
            const iconElement = container.querySelector('.icon-category');
            const iconcss = iconElement ? iconElement.className : '';
            // icon := "https://cdn.staticstack.net/includes/images/categories/" + iconbgcss + ".svg"
            const iconcssarr = iconcss.split(' ');
            let iconbgcss = '';
            for(let i=0;i<iconcssarr.length;i++){
                if(iconcssarr[i].startsWith('icon-category-')){
                    iconbgcss = iconcssarr[i];
                    break;
                }
            }
            iconbgcss = iconbgcss.replace('icon-category-', '');
            const iconurl = "https://cdn.staticstack.net/includes/images/categories/" + iconbgcss + ".svg";

            return {
                title: title,
                url: url,
                desc: desc,
                iconcss: iconcss,
                iconurl: iconurl
            };
        } catch (error) {
            console.error('Error extracting data from container:', error);
            return null;
        }
    }

    // 发送分类数据到API
    function sendCategoryToAPI(categoryData) {
        const postData = `class=${CONFIG.targetClass}&name=${encodeURIComponent(categoryData.title)}&url=${encodeURIComponent(categoryData.url)}&desc=${encodeURIComponent(categoryData.desc)}&iconcss=${encodeURIComponent(categoryData.iconcss)}&icon=${encodeURIComponent(categoryData.iconurl)}&is_used=1`;

        GM.xmlHttpRequest({
            method: 'POST',
            url: CONFIG.apiUrl,
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
                'Token': CONFIG.token
            },
            data: postData,
            onload: function(response) {
                console.log('API Response for category:', categoryData.title, '|', response.status, response.responseText);
            },
            onerror: function(error) {
                console.error('API Error for category:', categoryData.title, '|', error);
                showNotification(`API请求失败: ${categoryData.title}`, 'error');
            },
            ontimeout: function() {
                console.error('API Timeout for category:', categoryData.title);
                showNotification(`API请求超时: ${categoryData.title}`, 'warning');
            },
            timeout: 10000 // 10秒超时
        });
        // fetch(CONFIG.apiUrl, {
        //     method: 'POST',
        //     headers: {
        //         'Content-Type': 'application/x-www-form-urlencoded',
        //         'Token': CONFIG.token
        //     },
        //     body: postData,
        // }).then(response => {
        //     if (!response.ok) {
        //         console.error(`发送失败: ${response.statusText}`);
        //     } else {
        //         console.log(`已发送: ${title}`);
        //     }
        // }).catch(error => {
        //     console.error('请求错误:', error);
        // });
    }

    // 主函数
    async function getCategoriesAndSend() {
        try {
            // 等待页面加载完成
            await waitForPageLoad();

            // 额外等待一段时间，确保动态内容也加载完成
            setTimeout(() => {
                console.log('Starting category-container extraction...');

                // 查找所有category-container节点
                const containers = findCategoryContainers();

                if (containers.length === 0) {
                    console.log('No category-container elements found');
                    return;
                }

                // 遍历每个container
                containers.forEach((container, containerIndex) => {
                    console.log(`Processing container ${containerIndex + 1}/${containers.length}`);

                    // 提取分类数据
                    const categoryData = extractCategoryData(container);

                    if (!categoryData) {
                        console.log(`Failed to extract data from container ${containerIndex + 1}`);
                        return;
                    }

                    // 验证必要数据
                    if (!categoryData.title && !categoryData.url) {
                        console.log(`Container ${containerIndex + 1} has no valid title or URL, skipping`);
                        return;
                    }

                    // 输出提取的数据
                    console.log(`Category ${containerIndex + 1} data:`, {
                        title: categoryData.title,
                        url: categoryData.url,
                        desc: categoryData.desc,
                        iconcss: categoryData.iconcss
                    });

                    // 发送数据到API（添加延迟避免过于频繁的请求）
                    setTimeout(() => {
                        sendCategoryToAPI(categoryData);
                    }, containerIndex * 200); // 每个请求间隔200ms
                });

            }, 2000); // 页面加载完成后再等待2秒

        } catch (error) {
            console.error('Script error:', error);
        }
    }

    // 启动脚本
    getCategoriesAndSend();

    // 手动触发执行
    GM_registerMenuCommand('手动提取分类信息并发送', getCategoriesAndSend);

})();

// debug: 
// GM_xmlhttpRequest is not defined: @grant GM_xmlhttpRequest 或使用fetch代替
// 执行ok