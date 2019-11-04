$(document).ready(function () {
    $('.tab-content').find('.tab-pane').each(function (idx, item) {
        var navTabs = $(this).closest('.code-tabs').find('.nav-tabs'),
            title = $(this).attr('title');
        navTabs.append('<li><a href="#">' + title + '</a></li');
    });


    $('.nav-tabs a').click(function (e) {
        e.preventDefault();
        var tab = $(this).parent(),
            tabIndex = tab.index(),
            tabPanel = $(this).closest('.code-tabs'),
            tabPane = tabPanel.find('.tab-pane').eq(tabIndex);
        tabPanel.find('.active').removeClass('active');
        tab.addClass('active');
        tabPane.addClass('active');
        updateCurrentTab(tabPanel, tabIndex + 1)
    });

    function updateCurrentTab(panel, tabIndex) {
        panel.find('.active').removeClass('active');
        $('ul.nav-tabs', panel).find("li:nth-of-type(" + tabIndex + ")").addClass('active');
        $('.tab-content', panel).find("div:nth-of-type(" + tabIndex + ")").addClass('active');
    }

    $(".code-tabs").each(function () {
        updateCurrentTab($(this), 1);
    })
});
