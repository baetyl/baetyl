const log4js = require('log4js');
const moment = require('moment');

const getLogger = config => {
    if (!hasAttr(config, 'logger')) {
        return log4js.getLogger(config.name);
    }
    if (!hasAttr(config.logger, 'path')) {
        config.logger.path = 'var/log/' + config.name + '.log';
    }
    let level = 'info';
    if (hasAttr(config.logger, 'level')) {
        if (config.logger.level === 'debug') {
            level = 'debug';
        } else if (config.logger.level === 'warn') {
            level = 'warn';
        } else if (config.logger.level === 'error') {
            level = 'error';
        }
    }
    let backupCount = 15;
    if (hasAttr(config.logger, 'backupCount') && hasAttr(config.logger.backupCount, 'max')) {
        backupCount = config.logger.backupCount.max;
    }

    log4js.addLayout('openedge', () => logEvent => {
        const asctime = moment(logEvent.startTime).format('YYYY-MM-DD HH:mm:ss');
        const name = logEvent.categoryName;
        const levelname = logEvent.level.levelStr;
        const message = logEvent.data;
        return `${asctime} - ${name} - ${levelname} - ${message}`;
    });

    log4js.configure({
        appenders: {
            file: {
                type: 'file',
                filename: config.logger.path,
                layout: { type: 'openedge' },
                backups: backupCount,
                compress: true,
                encoding: 'utf-8'
            }
        },
        categories: {
            default: { appenders: ['file'], level }
        }
    });

    const logger = log4js.getLogger(config.name);
    return logger;
};


const hasAttr = (obj, attr) => {
    if (obj instanceof Object && !(obj instanceof Array)) {
        if (obj[attr]) {
            return true;
        }
    }
    return false;
};

class ConfigError {
    constructor(message) {
        this.message = message;
        this.name = 'ConfigError';
    }
}

module.exports = {
    getLogger,
    hasAttr,
    ConfigError
};