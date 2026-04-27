'use strict';

const { GLib, Gio } = imports.gi;
const Main = imports.ui.main;

const DBUS_INTERFACE = `
<node>
  <interface name="org.screenshooter.mcp.Windows">
    <method name="List">
      <arg type="aa{sv}" direction="out" name="windows"/>
    </method>
    <method name="Activate">
      <arg type="s" direction="in" name="title"/>
    </method>
    <method name="Move">
      <arg type="s" direction="in" name="title"/>
      <arg type="i" direction="in" name="x"/>
      <arg type="i" direction="in" name="y"/>
    </method>
    <method name="Resize">
      <arg type="s" direction="in" name="title"/>
      <arg type="i" direction="in" name="width"/>
      <arg type="i" direction="in" name="height"/>
    </method>
    <method name="ActiveTitle">
      <arg type="s" direction="out" name="title"/>
    </method>
    <method name="Close">
      <arg type="s" direction="in" name="title"/>
    </method>
    <method name="Minimize">
      <arg type="s" direction="in" name="title"/>
    </method>
    <method name="Maximize">
      <arg type="s" direction="in" name="title"/>
    </method>
  </interface>
</node>`;

var WindowsDBus = class WindowsDBus {
    _findWindow(title) {
        const lower = title.toLowerCase();
        return global.get_window_actors()
            .map(a => a.get_meta_window())
            .find(w => (w.get_title() || '').toLowerCase().includes(lower));
    }

    List() {
        const windows = global.get_window_actors()
            .filter(a => !a.get_meta_window().is_skip_taskbar())
            .map(a => {
                const w = a.get_meta_window();
                const r = w.get_frame_rect();
                return {
                    id:        new GLib.Variant('t', w.get_stable_sequence()),
                    title:     new GLib.Variant('s', w.get_title() || ''),
                    pid:       new GLib.Variant('i', w.get_pid()),
                    x:         new GLib.Variant('i', r.x),
                    y:         new GLib.Variant('i', r.y),
                    w:         new GLib.Variant('i', r.width),
                    h:         new GLib.Variant('i', r.height),
                    minimized: new GLib.Variant('b', w.minimized),
                    maximized: new GLib.Variant('b', w.maximized),
                };
            });
        return new GLib.Variant('(aa{sv})', [windows]);
    }

    ActiveTitle() {
        const w = global.display.get_focus_window();
        return new GLib.Variant('(s)', [w ? w.get_title() : '']);
    }

    Activate([title]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        w.activate(global.get_current_time());
    }

    Move([title, x, y]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        w.move_frame(true, x, y);
    }

    Resize([title, width, height]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        const r = w.get_frame_rect();
        w.move_resize_frame(true, r.x, r.y, width, height);
    }

    Close([title]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        w.delete(global.get_current_time());
    }

    Minimize([title]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        w.minimize();
    }

    Maximize([title]) {
        const w = this._findWindow(title);
        if (!w) throw new Error(`Window not found: ${title}`);
        w.maximize();
    }
};

var ScreenshooterMCPExtension = class ScreenshooterMCPExtension {
    constructor() {
        this._dbus = null;
        this._dbusId = null;
    }

    enable() {
        this._impl = new WindowsDBus();
        this._dbusId = Gio.DBus.session.own_name(
            'org.screenshooter.mcp',
            Gio.BusNameOwnerFlags.NONE,
            null, null
        );
        this._dbus = Gio.DBusExportedObject.wrapJSObject(
            DBUS_INTERFACE,
            this._impl
        );
        this._dbus.export(Gio.DBus.session, '/org/screenshooter/mcp');
    }

    disable() {
        if (this._dbus) {
            this._dbus.unexport();
            this._dbus = null;
        }
        if (this._dbusId) {
            Gio.DBus.session.unown_name(this._dbusId);
            this._dbusId = null;
        }
        this._impl = null;
    }
};

function init() {
    return new ScreenshooterMCPExtension();
}
