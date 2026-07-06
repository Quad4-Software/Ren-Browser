void window_apply_frameless(GtkWindow *window, gboolean frameless, const char *title) {
    if (!frameless) {
        gtk_window_set_titlebar(window, NULL);
        gtk_window_set_decorated(window, TRUE);
        if (title != NULL && title[0] != '\0') {
            gtk_window_set_title(window, title);
        }
        gtk_window_present(window);
        return;
    }
    gtk_window_set_decorated(window, FALSE);
}
