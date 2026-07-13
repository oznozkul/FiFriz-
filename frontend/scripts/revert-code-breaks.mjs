import fs from "fs";
import path from "path";

const srcDir = path.resolve("src");

const reverts = [
  ["seçtion", "section"],
  ["RecentAçtivitiesPanel", "RecentActivitiesPanel"],
  ["DashboardRecentAçtivity", "DashboardRecentActivity"],
  ["recentAçtivities", "recentActivities"],
  ["AçtivityPage", "ActivityPage"],
  ["RequireAuditAçcess", "RequireAuditAccess"],
  ["IntegrationAçcount", "IntegrationAccount"],
  ["isAçtive", "isActive"],
  ["setIsAçtive", "setIsActive"],
  ["setAçtiveTenant", "setActiveTenant"],
  ["setAçtiveMatchIndex", "setActiveMatchIndex"],
  ["setAçtiveSheetIndex", "setActiveSheetIndex"],
  ["setAçtiveTab", "setActiveTab"],
  ["setAçtingId", "setActingId"],
  ["setAçtingWorkflowId", "setActingWorkflowId"],
  ["setAçcountType", "setAccountType"],
  ["setAçtion", "setAction"],
  ["setAçcessToken", "setAccessToken"],
  ["getAçcessToken", "getAccessToken"],
  ["clearAçcessToken", "clearAccessToken"],
  ["canAçcess", "canAccess"],
  ["nextAçcessToken", "nextAccessToken"],
  ["auditAçtionLabels", "auditActionLabels"],
  ["AçcountType", "AccountType"],
  ["Çöpy", "Copy"],
  ["handleÇöpy", "handleCopy"],
  ["X-Share-Açcess-Token", "X-Share-Access-Token"],
  ["./AçtivityPage", "./ActivityPage"],
];

function walk(dir, files = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (entry.name !== "node_modules") walk(full, files);
    } else if (/\.(tsx?|ts)$/.test(entry.name)) {
      files.push(full);
    }
  }
  return files;
}

for (const file of walk(srcDir)) {
  let content = fs.readFileSync(file, "utf8");
  const original = content;
  for (const [from, to] of reverts) {
    if (content.includes(from)) {
      content = content.split(from).join(to);
    }
  }
  if (content !== original) {
    fs.writeFileSync(file, content, "utf8");
    console.log("fixed:", path.relative(srcDir, file));
  }
}
