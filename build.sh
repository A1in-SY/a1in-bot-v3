cd ./cmd
git pull
go build
mv cmd /home/admin/bot/
sudo systemctl restart bot
systemctl status bot