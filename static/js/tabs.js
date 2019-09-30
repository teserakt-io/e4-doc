$(document).ready(function () {
    $('.tab-content').find('.tab-pane').each(function (idx, item) {
        var navTabs = $(this).closest('.code-tabs').find('.nav-tabs'),
            title = $(this).attr('title');
        navTabs.append('<li><a href="#">' + title + '</a></li');
    });

    updateCurrentTab(1)

    $('.nav-tabs a').click(function (e) {
        e.preventDefault();
        var tab = $(this).parent(),
            tabIndex = tab.index(),
            tabPanel = $(this).closest('.code-tabs'),
            tabPane = tabPanel.find('.tab-pane').eq(tabIndex);
        tabPanel.find('.active').removeClass('active');
        tab.addClass('active');
        tabPane.addClass('active');

        updateCurrentTab(tabIndex + 1)
    });

    function updateCurrentTab(tabNumber) {
        $('.nav-tabs a').closest('.code-tabs').find('.active').removeClass('active');
        $('.code-tabs ul.nav-tabs').find("li:nth-of-type(" + tabNumber + ")").addClass('active');
        $('.code-tabs .tab-content').find("div:nth-of-type(" + tabNumber + ")").addClass('active');
    }
});
