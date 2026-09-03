let res = [
    db.createCollection("alarms"),
    db.createCollection("maintenance"),
    db.createCollection("notifications"),

    db.createRole({
        role: "roleApiUser",
        privileges: [
            {
                resource: { db: "unimq", collection: "acls" },
                actions: ["find", "insert", "update", "remove"],
            },
            {
                resource: { db: "unimq", collection: "alarms" },
                actions: ["find", "insert", "update", "remove"],
            },
            {
                resource: { db: "unimq", collection: "maintenance" },
                actions: ["find", "insert", "update", "remove"],
            },
            {
                resource: { db: "unimq", collection: "maintenance_edit_logs" },
                actions: ["find", "insert", "update", "remove"],
            },
            {
                resource: { db: "unimq", collection: "notifications" },
                actions: ["find", "insert", "update", "remove"],
            },
        ],
        roles: [{ role: "readWrite", db: "unimq" }],
    }),
];

printjson(res);
