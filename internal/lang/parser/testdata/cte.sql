with persons1 as (select firstname, lastname from employees) select firstname from persons1;

with persons2 as (
	select id, firstname, lastname from employees
), emails as (
	select id, mail from emails
)
select t1.firstname, t2.mail from persons2 t1 join emails t2 on t1.id=t2.id;